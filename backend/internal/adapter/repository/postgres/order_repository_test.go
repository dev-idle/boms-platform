package postgres_test

import (
	"context"
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	"github.com/boms/backend/internal/config"
	domaincart "github.com/boms/backend/internal/domain/cart"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderCheckoutRepositories_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Parallel()

	ctx := context.Background()
	pgContainer, connStr, err := startPostgres(ctx, t)
	if err != nil {
		if strings.Contains(err.Error(), "docker") {
			t.Skip("docker not available")
		}
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })
	require.NoError(t, applyMigrations(ctx, connStr))

	pool, err := postgresadapter.NewPool(ctx, config.PostgresConfig{
		URL:                connStr,
		MaxConns:           5,
		MinConns:           1,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute,
		HealthCheckTimeout: 5 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	users := postgresadapter.NewUserRepository(pool)
	categories := postgresadapter.NewCategoryRepository(pool)
	products := postgresadapter.NewProductRepository(pool)
	carts := postgresadapter.NewCartRepository(pool)
	orders := postgresadapter.NewOrderRepository(pool)

	customer, err := users.Create(ctx, port.CreateUserParams{
		Email:        "checkout@example.com",
		PasswordHash: testPasswordHashFixture,
		Role:         domainuser.RoleCustomer,
	})
	require.NoError(t, err)
	category, err := categories.Create(ctx, port.CreateCategoryParams{Name: "Cakes", Slug: "cakes", IsActive: true})
	require.NoError(t, err)
	cake, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: category.ID, Name: "Matcha cake", Slug: "matcha-cake", PriceCents: 4500, IsActive: true,
	})
	require.NoError(t, err)
	tart, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: category.ID, Name: "Fruit tart", Slug: "fruit-tart", PriceCents: 3200, IsActive: true,
	})
	require.NoError(t, err)
	_, err = carts.CreateForUser(ctx, customer.ID)
	require.NoError(t, err)

	t.Run("create_items_inserts_every_line_in_one_statement", func(t *testing.T) {
		order, err := orders.Create(ctx, port.CreateOrderParams{
			UserID: customer.ID, Status: domainorder.StatusPending, SubtotalCents: 12200, TotalCents: 12200,
		})
		require.NoError(t, err)

		require.NoError(t, orders.CreateItems(ctx, []port.CreateOrderItemParams{
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &cake.ID,
				Name: cake.Name, Slug: cake.Slug, Quantity: 2, UnitPriceCents: 4500, LineTotalCents: 9000,
			},
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &tart.ID,
				Configuration: json.RawMessage(`{"message":"Happy birthday"}`),
				Name:          tart.Name, Slug: tart.Slug, Quantity: 1, UnitPriceCents: 3200, LineTotalCents: 3200,
			},
		}))

		items, err := orders.ListItemsByOrderID(ctx, order.ID)
		require.NoError(t, err)
		require.Len(t, items, 2)
		byName := map[string]domainorder.Item{items[0].Name: items[0], items[1].Name: items[1]}
		assert.Equal(t, int32(2), byName[cake.Name].Quantity)
		assert.Equal(t, int64(9000), byName[cake.Name].LineTotalCents)
		assert.Nil(t, byName[cake.Name].ComboID)
		assert.JSONEq(t, `{}`, string(byName[cake.Name].Configuration))
		assert.JSONEq(t, `{"message":"Happy birthday"}`, string(byName[tart.Name].Configuration))
	})

	t.Run("create_items_is_atomic_when_a_line_violates_a_constraint", func(t *testing.T) {
		order, err := orders.Create(ctx, port.CreateOrderParams{
			UserID: customer.ID, Status: domainorder.StatusPending, SubtotalCents: 4500, TotalCents: 4500,
		})
		require.NoError(t, err)

		err = orders.CreateItems(ctx, []port.CreateOrderItemParams{
			{
				OrderID: order.ID, LineType: domaincart.LineTypeProduct, ProductID: &cake.ID,
				Name: cake.Name, Slug: cake.Slug, Quantity: 1, UnitPriceCents: 4500, LineTotalCents: 4500,
			},
			{
				// A product line without a product violates order_items_line_target_check.
				OrderID: order.ID, LineType: domaincart.LineTypeProduct,
				Name: "broken", Slug: "broken", Quantity: 1, UnitPriceCents: 0, LineTotalCents: 0,
			},
		})
		require.Error(t, err)

		items, err := orders.ListItemsByOrderID(ctx, order.ID)
		require.NoError(t, err)
		assert.Empty(t, items)
	})

	t.Run("get_for_update_requires_a_transaction", func(t *testing.T) {
		_, err := carts.GetByUserIDForUpdate(ctx, customer.ID)
		require.Error(t, err)
	})

	t.Run("get_for_update_serializes_concurrent_checkouts", func(t *testing.T) {
		const holdFor = 300 * time.Millisecond
		var firstReleased atomic.Bool
		locked := make(chan struct{})
		secondAcquired := make(chan bool, 1)

		go func() {
			<-locked
			err := pool.WithTx(ctx, func(txCtx context.Context) error {
				_, err := carts.GetByUserIDForUpdate(txCtx, customer.ID)
				secondAcquired <- firstReleased.Load()
				return err
			})
			if err != nil {
				secondAcquired <- false
			}
		}()

		require.NoError(t, pool.WithTx(ctx, func(txCtx context.Context) error {
			if _, err := carts.GetByUserIDForUpdate(txCtx, customer.ID); err != nil {
				return err
			}
			close(locked)
			time.Sleep(holdFor)
			firstReleased.Store(true)
			return nil
		}))

		select {
		case afterRelease := <-secondAcquired:
			assert.True(t, afterRelease, "second transaction acquired the cart lock while the first still held it")
		case <-time.After(10 * time.Second):
			t.Fatal("second transaction never acquired the cart lock")
		}
	})
}
