package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainsaved "github.com/boms/backend/internal/domain/saved"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// The engagement report counts each open customer account once for every
// feature it used, and how many used none, one, or two or more; a closed
// account does not count, whatever it used, and neither does staff.
func TestEngagement_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 8)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	saved := postgresadapter.NewSavedProductRepository(f.pool)
	conversations := postgresadapter.NewConversationRepository(f.pool)
	reviews := usecase.NewReviewUsecase(f.users, f.orders, postgresadapter.NewReviewRepository(f.pool), f.pool, f.outbox)
	report := usecase.NewManagerEngagementUsecase(postgresadapter.NewEngagementRepository(f.pool))
	paidOrder := func(t *testing.T, customer uuid.UUID, hour int) uuid.UUID {
		t.Helper()
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(hour, 0)))
		require.NoError(t, err)
		f.pay(t, customer, order.ID)
		return uuid.MustParse(order.ID)
	}
	write := func(t *testing.T, orderID, authorID uuid.UUID, role domainuser.Role, body string) {
		t.Helper()
		conversationID, err := conversations.OpenForCustomerMessage(ctx, orderID)
		require.NoError(t, err)
		_, err = conversations.CreateMessage(ctx, conversationID, authorID, role, body)
		require.NoError(t, err)
	}

	// A review and a favorite: two features.
	reviewer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
	pickedUp := paidOrder(t, reviewer, 9)
	_, err = raw.Exec(ctx, "UPDATE orders SET status = 'fulfilled' WHERE id = $1", pickedUp)
	require.NoError(t, err)
	_, err = reviews.Create(ctx, reviewer, pickedUp, dto.CreateReviewRequest{ProductID: f.pastry.String(), Rating: 5})
	require.NoError(t, err)
	_, err = saved.Save(ctx, reviewer, f.cake, domainsaved.ListFavorite)
	require.NoError(t, err)

	// A wishlist product taken off again and a message: two features; the
	// counter's reply is not the customer's.
	writer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
	_, err = saved.Save(ctx, writer, f.cake, domainsaved.ListWishlist)
	require.NoError(t, err)
	require.NoError(t, saved.Remove(ctx, writer, f.cake, domainsaved.ListWishlist))
	asked := paidOrder(t, writer, 10)
	write(t, asked, writer, domainuser.RoleCustomer, "Is it ready?")
	write(t, asked, f.newWorker(t, domainuser.RoleStaff), domainuser.RoleStaff, "Yes, at the counter.")

	// Promotion emails only: one feature.
	subscriber := f.newCustomer(t, nil, nil)
	agreed := true
	_, err = postgresadapter.NewCustomerProfileRepository(f.pool).Create(ctx, port.UpsertCustomerProfileParams{
		UserID: subscriber, MarketingOptIn: &agreed,
	})
	require.NoError(t, err)

	// Nothing used.
	f.newCustomer(t, nil, nil)

	// A closed account.
	closed := f.newCustomer(t, nil, nil)
	_, err = saved.Save(ctx, closed, f.cake, domainsaved.ListFavorite)
	require.NoError(t, err)
	require.NoError(t, f.pool.WithTx(ctx, func(txCtx context.Context) error { return f.users.Erase(txCtx, closed) }))

	out, err := report.Report(ctx)

	require.NoError(t, err)
	assert.Equal(t, &dto.EngagementReportResponse{
		Customers:     4,
		UsedNone:      1,
		UsedOne:       1,
		UsedTwoOrMore: 2,
		Features: []dto.EngagementFeatureResponse{
			{Feature: "reviews", Customers: 1},
			{Feature: "favorites", Customers: 1},
			{Feature: "wishlist", Customers: 1},
			{Feature: "messages", Customers: 1},
			{Feature: "promotions", Customers: 1},
		},
	}, out)
}
