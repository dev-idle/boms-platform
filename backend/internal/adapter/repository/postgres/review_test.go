package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainreview "github.com/boms/backend/internal/domain/review"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// A customer reviews each product they picked up once; a manager publishes or
// hides each review, the storefront shows only the published ones without
// their author, and the customer's export and erasure take the reviews along.
func TestReviews_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	reviewRepo := postgresadapter.NewReviewRepository(f.pool)
	reviews := usecase.NewReviewUsecase(f.users, f.orders, reviewRepo, f.pool, f.outbox)
	moderation := usecase.NewManagerReviewUsecase(reviewRepo, f.pool, f.outbox, nil, zap.NewNop())
	manager := f.newWorker(t, domainuser.RoleManager)
	_, err = postgresadapter.NewStaffProfileRepository(f.pool).Create(ctx, port.UpsertStaffProfileParams{
		UserID: manager, FullName: "Hoa Pham", EmployeeCode: "EMP-9101",
	})
	require.NoError(t, err)

	placed := func(t *testing.T, products, combos []uuid.UUID, hour int) (uuid.UUID, uuid.UUID) {
		t.Helper()
		customer := f.newCustomer(t, products, combos)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(hour, 0)))
		require.NoError(t, err)
		f.pay(t, customer, order.ID)
		return customer, uuid.MustParse(order.ID)
	}
	pickedUp := func(t *testing.T, products, combos []uuid.UUID, hour int) (uuid.UUID, uuid.UUID) {
		t.Helper()
		customer, orderID := placed(t, products, combos, hour)
		_, err := raw.Exec(ctx, "UPDATE orders SET status = 'fulfilled' WHERE id = $1", orderID)
		require.NoError(t, err)
		return customer, orderID
	}
	write := func(customer, orderID, productID uuid.UUID, rating int, comment *string) (*dto.ReviewResponse, error) {
		return reviews.Create(ctx, customer, orderID, dto.CreateReviewRequest{
			ProductID: productID.String(), Rating: rating, Comment: comment,
		})
	}
	moderate := func(t *testing.T, reviewID, status string) *dto.ManagerReviewResponse {
		t.Helper()
		out, err := moderation.Moderate(ctx, manager, domainuser.RoleManager, uuid.MustParse(reviewID), dto.ModerateReviewRequest{Status: status})
		require.NoError(t, err)
		return out
	}
	toAuthorAndManagers := func(author uuid.UUID) domainevent.Audience {
		return domainevent.Audience{UserIDs: []uuid.UUID{author}, Roles: []domainuser.Role{domainuser.RoleManager}}
	}

	t.Run("each_product_picked_up_is_reviewed_once", func(t *testing.T) {
		customer, orderID := pickedUp(t, []uuid.UUID{f.pastry, f.cake}, nil, 9)

		flaky, err := write(customer, orderID, f.pastry, 5, new("  Flaky and buttery\r\nWill order again  "))
		require.NoError(t, err)
		assert.Equal(t, "Flaky and buttery\nWill order again", *flaky.Comment)
		assert.Equal(t, string(domainreview.StatusPending), flaky.Status, "nothing shows before a manager reads it")
		rated, err := write(customer, orderID, f.cake, 4, new("   "))
		require.NoError(t, err)
		assert.Nil(t, rated.Comment, "a rating alone is a review")

		_, err = write(customer, orderID, f.pastry, 3, nil)
		require.ErrorIs(t, err, domainreview.ErrExists)
		_, err = write(customer, orderID, uuid.New(), 3, nil)
		require.ErrorIs(t, err, domainreview.ErrNotReviewable, "only a product on the order")

		mine, err := reviews.ListByOrder(ctx, customer, orderID)
		require.NoError(t, err)
		require.Len(t, mine, 2)
		assert.Equal(t, flaky.ID, mine[0].ID)
		assert.Equal(t, []domainevent.Audience{toAuthorAndManagers(customer)}, reviewAudiences(t, f, f.pastry),
			"managers hear of a review waiting, and the author's other pages")
	})

	t.Run("only_a_product_line_picked_up_by_its_customer_is_reviewed", func(t *testing.T) {
		customer, waiting := placed(t, []uuid.UUID{f.pastry}, nil, 10)
		_, err := write(customer, waiting, f.pastry, 5, nil)
		require.ErrorIs(t, err, domainreview.ErrNotReviewable, "not before it is picked up")

		comboCustomer, comboOrder := pickedUp(t, nil, []uuid.UUID{f.combo}, 11)
		_, err = write(comboCustomer, comboOrder, f.pastry, 5, nil)
		require.ErrorIs(t, err, domainreview.ErrNotReviewable, "a combo is not one of its products")

		stranger := f.newCustomer(t, nil, nil)
		_, err = write(stranger, comboOrder, f.pastry, 5, nil)
		require.ErrorIs(t, err, domainorder.ErrNotFound)
	})

	t.Run("the_storefront_shows_what_a_manager_published", func(t *testing.T) {
		tart := newSavedTestProduct(t, raw, f.checkoutFixture, "Lemon tart")
		var written []*dto.ReviewResponse
		var authors []uuid.UUID
		for i, rating := range []int{2, 4, 5} {
			customer, orderID := pickedUp(t, []uuid.UUID{tart}, nil, 12+i)
			review, err := write(customer, orderID, tart, rating, new(fmt.Sprintf("%d stars", rating)))
			require.NoError(t, err)
			written = append(written, review)
			authors = append(authors, customer)
		}
		before, err := reviews.ProductReviews(ctx, tart, nil)
		require.NoError(t, err)
		assert.Zero(t, before.ReviewCount)
		assert.Nil(t, before.AverageRating)
		assert.Empty(t, before.Reviews)

		hidden := moderate(t, written[0].ID, "hidden")
		assert.Equal(t, "Hoa Pham", *hidden.ModeratorName)
		moderate(t, written[1].ID, "published")
		moderate(t, written[2].ID, "published")
		assert.Equal(t, []domainevent.Audience{
			toAuthorAndManagers(authors[0]), toAuthorAndManagers(authors[1]), toAuthorAndManagers(authors[2]),
			toAuthorAndManagers(authors[0]), toAuthorAndManagers(authors[1]), toAuthorAndManagers(authors[2]),
		}, reviewAudiences(t, f, tart), "a review read tells managers and its author, as one written does")

		shown, err := reviews.ProductReviews(ctx, tart, nil)
		require.NoError(t, err)
		assert.Equal(t, int64(2), shown.ReviewCount)
		assert.InDelta(t, 4.5, *shown.AverageRating, 0)
		require.Len(t, shown.Reviews, 2)
		assert.Equal(t, written[2].ID, shown.Reviews[0].ID, "latest first")
		assert.False(t, shown.HasMore)

		summary, err := moderation.Summary(ctx)
		require.NoError(t, err)
		var rated *dto.ProductRatingResponse
		for i := range summary.Products {
			if summary.Products[i].ProductID == tart.String() {
				rated = &summary.Products[i]
			}
		}
		require.NotNil(t, rated)
		assert.Equal(t, int64(2), rated.ReviewCount, "a hidden review does not count")
		assert.InDelta(t, 4.5, rated.AverageRating, 0)
		assert.Len(t, summary.Stars, domainreview.MaxRating)
		assert.GreaterOrEqual(t, summary.PendingCount, int64(2), "the reviews written first still wait")

		hiddenOnes, _, _, _, err := moderation.List(ctx, 1, 100, "hidden")
		require.NoError(t, err)
		require.NotEmpty(t, hiddenOnes)
		assert.Equal(t, written[0].ID, hiddenOnes[0].ID)
		assert.Equal(t, "Lemon tart", hiddenOnes[0].ProductName)
		_, _, _, _, err = moderation.List(ctx, 1, 100, "deleted")
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "validation_error", appErr.Code)

		_, err = moderation.Moderate(ctx, manager, domainuser.RoleManager, uuid.MustParse(written[1].ID), dto.ModerateReviewRequest{Status: "pending"})
		require.ErrorIs(t, err, domainreview.ErrInvalidStatus, "a review read is never pending again")
		_, err = moderation.Moderate(ctx, manager, domainuser.RoleManager, uuid.New(), dto.ModerateReviewRequest{Status: "hidden"})
		require.ErrorIs(t, err, domainreview.ErrNotFound)

		_, err = raw.Exec(ctx, "UPDATE products SET deleted_at = now() WHERE id = $1", tart)
		require.NoError(t, err)
		summary, err = moderation.Summary(ctx)
		require.NoError(t, err)
		for _, product := range summary.Products {
			assert.NotEqual(t, tart.String(), product.ProductID, "a product the bakery deleted needs no attention")
		}
	})

	t.Run("a_product_s_reviews_page_back_from_the_last_shown", func(t *testing.T) {
		cake := newSavedTestProduct(t, raw, f.checkoutFixture, "Plum cake")
		for i := range 11 {
			customer, orderID := pickedUp(t, []uuid.UUID{cake}, nil, 9+i%5)
			review, err := write(customer, orderID, cake, 5, nil)
			require.NoError(t, err)
			moderate(t, review.ID, "published")
		}

		latest, err := reviews.ProductReviews(ctx, cake, nil)
		require.NoError(t, err)
		require.Len(t, latest.Reviews, 10)
		assert.True(t, latest.HasMore)
		assert.Equal(t, int64(11), latest.ReviewCount)

		older, err := reviews.ProductReviews(ctx, cake, new(uuid.MustParse(latest.Reviews[9].ID)))
		require.NoError(t, err)
		require.Len(t, older.Reviews, 1)
		assert.False(t, older.HasMore)
		for _, shown := range latest.Reviews {
			assert.NotEqual(t, shown.ID, older.Reviews[0].ID, "no review shows twice")
		}
	})

	t.Run("the_export_carries_the_reviews_and_erasure_takes_them", func(t *testing.T) {
		customer, orderID := pickedUp(t, []uuid.UUID{f.pastry}, nil, 16)
		review, err := write(customer, orderID, f.pastry, 3, new("A little dry"))
		require.NoError(t, err)
		moderate(t, review.ID, "published")

		customerProfiles := postgresadapter.NewCustomerProfileRepository(f.pool)
		audit := postgresadapter.NewAuditLogRepository(f.pool)
		conversations := postgresadapter.NewConversationRepository(f.pool)
		savedRepo := postgresadapter.NewSavedProductRepository(f.pool)
		export, err := usecase.NewDataExportUsecase(f.users, customerProfiles, postgresadapter.NewStaffProfileRepository(f.pool),
			postgresadapter.NewAdminProfileRepository(f.pool), f.orders, conversations, savedRepo, reviewRepo, f.carts, fixedSessions{}, audit,
		).Export(ctx, customer)
		require.NoError(t, err)
		require.Len(t, export.Reviews, 1)
		assert.Equal(t, "A little dry", *export.Reviews[0].Comment)
		assert.Equal(t, "published", export.Reviews[0].Status)

		erasure := usecase.NewAccountErasureUsecase(f.pool, f.users, customerProfiles, f.carts, f.orders, conversations, savedRepo, reviewRepo, audit,
			postgresadapter.NewUserTokenRepository(f.pool), &endedSessions{}, auditlogger.NewService(audit), fixtureHasher{})
		require.NoError(t, erasure.Erase(ctx, customer, fixturePassword))

		left, err := reviews.ListByOrder(ctx, customer, orderID)
		require.NoError(t, err)
		assert.Empty(t, left)
		var comment *string
		require.NoError(t, raw.QueryRow(ctx, "SELECT comment FROM reviews WHERE id = $1", review.ID).Scan(&comment))
		assert.Nil(t, comment, "their text is gone, not only hidden")
	})
}

// reviewAudiences lists who the outbox's review events about a product reach,
// oldest first.
func reviewAudiences(t *testing.T, f *ticketFixture, productID uuid.UUID) []domainevent.Audience {
	t.Helper()
	var events []domainevent.Event
	require.NoError(t, f.pool.WithTx(context.Background(), func(txCtx context.Context) error {
		var err error
		events, err = f.outbox.ClaimUnpublished(txCtx, 0, 1000)
		return err
	}))
	var audiences []domainevent.Audience
	for _, e := range events {
		if e.Topic == domainreview.TopicChanged && e.Data["product_id"] == productID.String() {
			audiences = append(audiences, e.Audience)
		}
	}
	return audiences
}
