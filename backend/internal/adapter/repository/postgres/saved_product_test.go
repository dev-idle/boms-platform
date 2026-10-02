package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainsaved "github.com/boms/backend/internal/domain/saved"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/usecase"
)

// A customer keeps products on their favorites and their wishlist: a product
// saved again keeps its place, one taken off comes back on top, one the
// bakery stopped selling is kept out of sight until it is back, and the lists
// go with the account's export and erasure.
func TestSavedProducts_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 8)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	savedRepo := postgresadapter.NewSavedProductRepository(f.pool)
	saved := usecase.NewSavedProductUsecase(f.users, savedRepo, f.pool)

	listed := func(t *testing.T, customer uuid.UUID, list domainsaved.List) []dto.SavedProductResponse {
		t.Helper()
		rows, err := saved.List(ctx, customer)
		require.NoError(t, err)
		var out []dto.SavedProductResponse
		for _, row := range rows {
			if row.List == string(list) {
				out = append(out, row)
			}
		}
		return out
	}
	names := func(rows []dto.SavedProductResponse) []string {
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, row.Product.Name)
		}
		return out
	}

	t.Run("each_list_keeps_its_own_products_latest_first", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		require.NoError(t, saved.Save(ctx, customer, "favorite", f.cake))
		require.NoError(t, saved.Save(ctx, customer, "favorite", f.pastry))
		require.NoError(t, saved.Save(ctx, customer, "wishlist", f.cake))

		favorites := listed(t, customer, domainsaved.ListFavorite)
		assert.Equal(t, []string{"Croissant", "Matcha cake"}, names(favorites))
		assert.Equal(t, int64(300), favorites[0].Product.PriceCents, "as the catalog shows it")
		assert.Equal(t, []string{"Matcha cake"}, names(listed(t, customer, domainsaved.ListWishlist)))

		require.NoError(t, saved.Save(ctx, customer, "favorite", f.cake))
		assert.Equal(t, []string{"Croissant", "Matcha cake"}, names(listed(t, customer, domainsaved.ListFavorite)),
			"saving it again leaves it where it was")

		require.NoError(t, saved.Remove(ctx, customer, "favorite", f.cake))
		require.NoError(t, saved.Remove(ctx, customer, "favorite", f.cake), "taking it off twice is no error")
		assert.Equal(t, []string{"Croissant"}, names(listed(t, customer, domainsaved.ListFavorite)))
		assert.Equal(t, []string{"Matcha cake"}, names(listed(t, customer, domainsaved.ListWishlist)), "the other list keeps it")

		require.NoError(t, saved.Save(ctx, customer, "favorite", f.cake))
		assert.Equal(t, []string{"Matcha cake", "Croissant"}, names(listed(t, customer, domainsaved.ListFavorite)),
			"back on the list, it is the latest")
	})

	t.Run("only_a_product_on_sale_is_saved_or_shown", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		require.ErrorIs(t, saved.Save(ctx, customer, "favorite", uuid.New()), domainproduct.ErrNotFound)
		require.ErrorIs(t, saved.Save(ctx, customer, "basket", f.cake), domainsaved.ErrInvalidList)

		withdrawn := newSavedTestProduct(t, raw, f, "Plum tart")
		require.NoError(t, saved.Save(ctx, customer, "wishlist", withdrawn))
		_, err := raw.Exec(ctx, `UPDATE products SET is_active = false WHERE id = $1`, withdrawn)
		require.NoError(t, err)

		assert.Empty(t, listed(t, customer, domainsaved.ListWishlist), "nothing of it shows while it is not sold")
		require.ErrorIs(t, saved.Save(ctx, customer, "favorite", withdrawn), domainproduct.ErrNotFound)
		held, err := savedRepo.ListForExport(ctx, customer)
		require.NoError(t, err)
		require.Len(t, held, 1, "the list still holds it")

		_, err = raw.Exec(ctx, `UPDATE products SET is_active = true WHERE id = $1`, withdrawn)
		require.NoError(t, err)
		assert.Equal(t, []string{"Plum tart"}, names(listed(t, customer, domainsaved.ListWishlist)), "back on sale, back on the list")
	})

	t.Run("a_full_list_takes_no_more_but_keeps_what_it_holds", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, err := raw.Exec(ctx, `
			WITH made AS (
				INSERT INTO products (category_id, name, slug, price_cents, is_active)
				SELECT p.category_id, 'Bun ' || n || ' ' || $2::text, 'bun-' || n || '-' || $2::text, 100, true
				FROM products p, generate_series(1, $3::int) n
				WHERE p.id = $1
				RETURNING id
			)
			INSERT INTO saved_products (user_id, product_id, list)
			SELECT $4, id, 'favorite' FROM made`,
			f.pastry, customer.String()[:8], domainsaved.MaxPerList, customer)
		require.NoError(t, err)

		require.ErrorIs(t, saved.Save(ctx, customer, "favorite", f.cake), domainsaved.ErrListFull)
		some := listed(t, customer, domainsaved.ListFavorite)[0].Product.ID
		require.NoError(t, saved.Save(ctx, customer, "favorite", uuid.MustParse(some)), "one already on it is no new product")
		require.NoError(t, saved.Save(ctx, customer, "wishlist", f.cake), "the other list has room")

		_, err = raw.Exec(ctx, `UPDATE products SET is_active = false WHERE id = $1`, some)
		require.NoError(t, err)
		require.NoError(t, saved.Save(ctx, customer, "favorite", f.cake), "a product not shown takes no place")
	})

	t.Run("the_export_carries_the_lists_and_erasure_empties_them", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		require.NoError(t, saved.Save(ctx, customer, "favorite", f.pastry))
		require.NoError(t, saved.Save(ctx, customer, "wishlist", f.cake))
		require.NoError(t, saved.Save(ctx, customer, "favorite", f.cake))
		require.NoError(t, saved.Remove(ctx, customer, "favorite", f.cake))
		customerProfiles := postgresadapter.NewCustomerProfileRepository(f.pool)
		audit := postgresadapter.NewAuditLogRepository(f.pool)
		conversations := postgresadapter.NewConversationRepository(f.pool)

		export, err := usecase.NewDataExportUsecase(f.users, customerProfiles, postgresadapter.NewStaffProfileRepository(f.pool),
			postgresadapter.NewAdminProfileRepository(f.pool), f.orders, conversations, savedRepo, postgresadapter.NewReviewRepository(f.pool), f.carts, fixedSessions{}, audit,
		).Export(ctx, customer)
		require.NoError(t, err)
		held := make([]string, 0, len(export.Saved))
		for _, item := range export.Saved {
			held = append(held, fmt.Sprintf("%s:%s:%t", item.List, item.ProductName, item.RemovedAt != nil))
		}
		assert.ElementsMatch(t, []string{"favorite:Croissant:false", "favorite:Matcha cake:true", "wishlist:Matcha cake:false"}, held,
			"a product taken off a list is still held, with when")

		erasure := usecase.NewAccountErasureUsecase(f.pool, f.users, customerProfiles, f.carts, f.orders, conversations, savedRepo, postgresadapter.NewReviewRepository(f.pool), audit,
			postgresadapter.NewUserTokenRepository(f.pool), &endedSessions{}, auditlogger.NewService(audit), fixtureHasher{})
		require.NoError(t, erasure.Erase(ctx, customer, fixturePassword))

		left, err := savedRepo.ListByUser(ctx, customer)
		require.NoError(t, err)
		assert.Empty(t, left)
		require.ErrorIs(t, saved.Save(ctx, customer, "favorite", f.cake), usecase.ErrMeNotFound, "a closed account saves nothing")
	})
}

// newSavedTestProduct adds a product on sale in the pastry's category.
func newSavedTestProduct(t *testing.T, raw *pgxpool.Pool, f *checkoutFixture, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	require.NoError(t, raw.QueryRow(context.Background(), `
		INSERT INTO products (category_id, name, slug, price_cents, is_active)
		SELECT category_id, $2, $3, 450, true FROM products WHERE id = $1
		RETURNING id`, f.pastry, name, fmt.Sprintf("saved-%s", uuid.NewString()[:8])).Scan(&id))
	return id
}
