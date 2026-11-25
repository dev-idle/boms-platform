package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
)

// The admin home counts accounts by role, the disabled ones, customers new
// this week and those yet to confirm their email, events the outbox has held
// too long and refunds left unmade.
func TestAdminDashboardRepository_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	dashboard := postgresadapter.NewAdminDashboardRepository(f.pool)
	read := func() (map[domainuser.Role]int64, port.AdminDashboardCounts) {
		roles, err := dashboard.CountActiveUsersByRole(ctx)
		require.NoError(t, err)
		counts, err := dashboard.Counts(ctx, time.Now().Add(-time.Hour), 5*time.Minute, time.Hour)
		require.NoError(t, err)
		return roles, counts
	}
	rolesBefore, before := read()

	_, err = f.users.Create(ctx, port.CreateUserParams{
		Email: "unconfirmed@example.com", PasswordHash: testPasswordHashFixture, Role: domainuser.RoleCustomer,
	})
	require.NoError(t, err)
	baker := f.newWorker(t, domainuser.RoleBaker)
	require.NoError(t, f.users.SoftDelete(ctx, baker))
	erased := f.newCustomer(t, nil, nil)
	require.NoError(t, f.pool.WithTx(ctx, func(txCtx context.Context) error { return f.users.Erase(txCtx, erased) }))
	_, err = raw.Exec(ctx, `INSERT INTO outbox_events (id, topic, audience, data, created_at)
		VALUES ($1, 'order.status_changed', '{}', '{}', now() - interval '10 minutes'),
		       ($2, 'order.status_changed', '{}', '{}', now())`, uuid.New(), uuid.New())
	require.NoError(t, err)
	paid := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 10)
	_, err = raw.Exec(ctx, "UPDATE payments SET refund_requested_at = now() - interval '2 hours' WHERE order_id = $1", paid.ID)
	require.NoError(t, err)

	rolesAfter, after := read()

	assert.Equal(t, rolesBefore[domainuser.RoleCustomer]+2, rolesAfter[domainuser.RoleCustomer], "the new customer and the one who paid")
	assert.Equal(t, rolesBefore[domainuser.RoleBaker], rolesAfter[domainuser.RoleBaker], "a disabled account is not in use")
	assert.Equal(t, before.Disabled+1, after.Disabled, "an erased account is gone, not disabled")
	assert.Equal(t, before.NewCustomers+2, after.NewCustomers)
	assert.Equal(t, before.UnconfirmedEmails+1, after.UnconfirmedEmails, "the customer who paid confirmed their email")
	assert.Equal(t, before.UndeliveredEvents+1, after.UndeliveredEvents, "an event just written is on its way")
	assert.Equal(t, before.StuckRefunds+1, after.StuckRefunds)
}
