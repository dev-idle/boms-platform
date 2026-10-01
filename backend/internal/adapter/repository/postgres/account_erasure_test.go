package postgres_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainaccount "github.com/boms/backend/internal/domain/account"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// endedSessions stands in for the Redis session store and records whose
// sessions were ended; any other call panics.
type endedSessions struct {
	port.SessionStore
	ended []string
}

func (s *endedSessions) DeleteAllForUser(_ context.Context, userID string) error {
	s.ended = append(s.ended, userID)
	return nil
}

// fixturePassword is the password behind testPasswordHashFixture.
const fixturePassword = "the-fixture-password"

// fixtureHasher accepts fixturePassword for the fixture hash, like Argon2 would
// for the password a real hash was made from.
type fixtureHasher struct{ port.PasswordHasher }

func (fixtureHasher) Verify(encoded, password string) error {
	if encoded != testPasswordHashFixture || password != fixturePassword {
		return errors.New("password mismatch")
	}
	return nil
}

// A customer who asks to be forgotten is: their details go, their orders stay
// as anonymous sales records, and nobody can bring the account back.
func TestAccountErasure_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 8)
	customerProfiles := postgresadapter.NewCustomerProfileRepository(f.pool)
	audit := postgresadapter.NewAuditLogRepository(f.pool)
	sessions := &endedSessions{}
	tokens := postgresadapter.NewUserTokenRepository(f.pool)
	erasure := usecase.NewAccountErasureUsecase(
		f.pool, f.users, customerProfiles, f.carts, f.orders, audit, tokens, sessions,
		auditlogger.NewService(audit), fixtureHasher{},
	)

	t.Run("erases_the_details_and_keeps_the_orders", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		account, err := f.users.AdminGetByID(ctx, customer)
		require.NoError(t, err)
		name, phone := "Mai", "+84907654321"
		_, err = customerProfiles.Create(ctx, port.UpsertCustomerProfileParams{UserID: customer, DisplayName: &name, Phone: &phone})
		require.NoError(t, err)
		collected, err := f.orders.Create(ctx, port.CreateOrderParams{
			UserID: &customer, Channel: domainorder.ChannelOnline, Code: "CH-250102-001", Status: domainorder.StatusFulfilled,
			Type: domainorder.TypeInstant, SubtotalCents: 300, TotalCents: 300,
		})
		require.NoError(t, err)
		admin := f.newWorker(t, domainuser.RoleAdmin)
		ownIP, adminIP := "198.51.100.4", "192.0.2.9"
		require.NoError(t, audit.Create(ctx, port.CreateAuditLogParams{
			ActorID: customer, ActorRole: domainuser.RoleCustomer, Action: domainuser.AuditActionMeUpdatedProfile,
			TargetID: &customer, TargetType: "user_profile", BeforeJSON: []byte(`{"phone":null}`), AfterJSON: []byte(`{"phone":"+84907654321"}`), IP: &ownIP,
		}))
		require.NoError(t, audit.Create(ctx, port.CreateAuditLogParams{
			ActorID: admin, ActorRole: domainuser.RoleAdmin, Action: domainuser.AuditActionAdminEnabledUser,
			TargetID: &customer, TargetType: "user", BeforeJSON: []byte(`{"disabled":true}`), AfterJSON: []byte(`{"phone_released":"+84907654321"}`), IP: &adminIP,
		}))

		require.ErrorIs(t, erasure.Erase(ctx, customer, "not-the-password"), apperrors.ErrInvalidCredentials)
		require.NoError(t, erasure.Erase(ctx, customer, fixturePassword))

		erased, err := f.users.AdminGetByID(ctx, customer)
		require.NoError(t, err)
		assert.True(t, erased.Disabled(), "the account is closed")
		assert.True(t, erased.Erased())
		assert.True(t, strings.HasSuffix(erased.Email, "@erased.invalid"), "the email is gone: %s", erased.Email)

		profile, err := customerProfiles.GetByUserID(ctx, customer)
		require.NoError(t, err)
		assert.Nil(t, profile.DisplayName)
		assert.Nil(t, profile.Phone)

		cart, err := f.carts.GetByUserID(ctx, customer)
		require.NoError(t, err)
		items, err := f.carts.ListItemsByCartID(ctx, cart.ID)
		require.NoError(t, err)
		assert.Empty(t, items, "the cart is empty")

		kept, err := f.orders.GetByIDForUser(ctx, customer, collected.ID)
		require.NoError(t, err)
		assert.Equal(t, collected.Code, kept.Code, "the sale stays on record")

		trail, err := audit.ListForSubject(ctx, customer, nil, 100)
		require.NoError(t, err)
		require.Len(t, trail, 3, "the erasure, the admin's change and the customer's own")
		for _, entry := range trail {
			assert.JSONEq(t, `{}`, string(entry.AfterJSON), "%s keeps no personal data", entry.Action)
			if entry.ActorID == customer {
				assert.Empty(t, entry.IP, "%s keeps no network of the customer", entry.Action)
				assert.Nil(t, entry.UserAgent)
			}
		}
		assert.Equal(t, domainuser.AuditActionMeErasedAccount, trail[0].Action)
		assert.Equal(t, adminIP, trail[1].IP, "the admin's own network is theirs to keep")
		assert.Contains(t, sessions.ended, customer.String(), "every session ends")

		require.ErrorIs(t, f.users.Restore(ctx, customer), apperrors.ErrNotFound, "an erased account never comes back")
		_, err = f.users.Create(ctx, port.CreateUserParams{
			Email: account.Email, PasswordHash: testPasswordHashFixture, Role: domainuser.RoleCustomer,
		})
		require.NoError(t, err, "the address is free for a new account")
	})

	t.Run("waits_until_no_order_is_open", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		_, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(12, 0)))
		require.NoError(t, err)
		link, err := tokens.Issue(ctx, customer, domainaccount.PurposeResetPassword)
		require.NoError(t, err)

		require.ErrorIs(t, erasure.Erase(ctx, customer, fixturePassword), domainuser.ErrAccountHasOpenOrders)

		account, err := f.users.AdminGetByID(ctx, customer)
		require.NoError(t, err)
		assert.False(t, account.Disabled(), "nothing was erased")
		assert.NotContains(t, sessions.ended, customer.String())
		owner, err := tokens.Redeem(ctx, link, domainaccount.PurposeResetPassword)
		require.NoError(t, err, "the account's links still work")
		assert.Equal(t, customer, owner)
	})

	t.Run("asking_again_only_ends_the_sessions", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		require.NoError(t, erasure.Erase(ctx, customer, fixturePassword))
		ended := len(sessions.ended)

		require.NoError(t, erasure.Erase(ctx, customer, ""), "a retry after a failed sign-out finishes it")
		assert.Len(t, sessions.ended, ended+1)

		disabled := f.newCustomer(t, nil, nil)
		require.NoError(t, f.users.SoftDelete(ctx, disabled))
		require.ErrorIs(t, erasure.Erase(ctx, disabled, fixturePassword), usecase.ErrMeNotFound, "a disabled account is not erased")
	})

	t.Run("an_erased_account_cannot_check_out", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		require.NoError(t, erasure.Erase(ctx, customer, fixturePassword))
		// A session that outlived the erasure fills the cart again.
		require.NoError(t, f.fillCart(customer, []uuid.UUID{f.pastry}, nil))

		_, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(12, 0)))

		require.ErrorIs(t, err, usecase.ErrMeNotFound)
	})

	t.Run("erases_a_profile_change_it_waited_for", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		_, err := customerProfiles.Create(ctx, port.UpsertCustomerProfileParams{UserID: customer})
		require.NoError(t, err)
		holding := make(chan struct{})
		write := make(chan struct{})
		written := make(chan error, 1)
		// A PATCH /me that has taken its share of the account and is about to write.
		go func() {
			written <- f.pool.WithTx(ctx, func(txCtx context.Context) error {
				if _, err := f.users.GetByIDForShare(txCtx, customer); err != nil {
					return err
				}
				close(holding)
				<-write
				name, ip := "Mai", "198.51.100.7"
				if _, err := customerProfiles.UpdateByUserID(txCtx, port.UpsertCustomerProfileParams{UserID: customer, DisplayName: &name}); err != nil {
					return err
				}
				return audit.Create(txCtx, port.CreateAuditLogParams{
					ActorID: customer, ActorRole: domainuser.RoleCustomer, Action: domainuser.AuditActionMeUpdatedProfile,
					TargetID: &customer, TargetType: "user_profile", BeforeJSON: []byte(`{}`), AfterJSON: []byte(`{"display_name":"Mai"}`), IP: &ip,
				})
			})
		}()
		<-holding
		erased := make(chan error, 1)
		go func() { erased <- erasure.Erase(ctx, customer, fixturePassword) }()

		select {
		case err := <-erased:
			t.Fatalf("the erasure did not wait for the profile change: %v", err)
		case <-time.After(300 * time.Millisecond):
		}
		close(write)

		require.NoError(t, <-written)
		require.NoError(t, <-erased)
		profile, err := customerProfiles.GetByUserID(ctx, customer)
		require.NoError(t, err)
		assert.Nil(t, profile.DisplayName, "the change is erased with the rest")
		trail, err := audit.ListForSubject(ctx, customer, nil, 100)
		require.NoError(t, err)
		require.Len(t, trail, 2)
		assert.Equal(t, domainuser.AuditActionMeUpdatedProfile, trail[1].Action)
		assert.JSONEq(t, `{}`, string(trail[1].AfterJSON), "its record is scrubbed too")
		assert.Empty(t, trail[1].IP)
	})

	t.Run("waits_for_an_order_being_placed", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		holding := make(chan struct{})
		place := make(chan struct{})
		placed := make(chan error, 1)
		// A checkout that has taken its share of the account and is about to insert.
		go func() {
			placed <- f.pool.WithTx(ctx, func(txCtx context.Context) error {
				if _, err := f.users.GetByIDForShare(txCtx, customer); err != nil {
					return err
				}
				close(holding)
				<-place
				_, err := f.orders.Create(txCtx, port.CreateOrderParams{
					UserID: &customer, Channel: domainorder.ChannelOnline, Code: "CH-250103-001", Status: domainorder.StatusPending,
					Type: domainorder.TypeInstant, SubtotalCents: 300, TotalCents: 300,
				})
				return err
			})
		}()
		<-holding
		erased := make(chan error, 1)
		go func() { erased <- erasure.Erase(ctx, customer, fixturePassword) }()

		select {
		case err := <-erased:
			t.Fatalf("the erasure did not wait for the order being placed: %v", err)
		case <-time.After(300 * time.Millisecond):
		}
		close(place)

		require.NoError(t, <-placed)
		require.ErrorIs(t, <-erased, domainuser.ErrAccountHasOpenOrders, "the erasure sees the order it waited for")
	})
}
