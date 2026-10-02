package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainsession "github.com/boms/backend/internal/domain/session"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// fixedSessions stands in for the Redis session store: every user is signed in
// on the sessions given.
type fixedSessions []domainsession.SessionMeta

func (s fixedSessions) ListForUser(context.Context, string) ([]domainsession.SessionMeta, error) {
	return s, nil
}

// A person's acceptance of the policies is recorded with the account or order
// it came with, and a data export hands back everything held about them.
func TestPolicyAcceptanceAndDataExport_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 8)
	customerProfiles := postgresadapter.NewCustomerProfileRepository(f.pool)
	audit := postgresadapter.NewAuditLogRepository(f.pool)
	sessions := fixedSessions{{
		RefreshJTI: "refresh-id-that-stays-private", CreatedAt: time.Now().UTC(), IP: "203.0.113.7", UserAgent: "Test browser",
	}}
	exports := usecase.NewDataExportUsecase(
		f.users,
		customerProfiles,
		postgresadapter.NewStaffProfileRepository(f.pool),
		postgresadapter.NewAdminProfileRepository(f.pool),
		f.orders,
		postgresadapter.NewConversationRepository(f.pool),
		postgresadapter.NewSavedProductRepository(f.pool),
		postgresadapter.NewReviewRepository(f.pool),
		f.carts,
		sessions,
		audit,
	)

	t.Run("an_account_records_the_version_accepted_at_sign_up", func(t *testing.T) {
		version := domainpolicy.TermsVersion
		signedUp, err := f.users.Create(ctx, port.CreateUserParams{
			Email: "accepted@example.com", PasswordHash: testPasswordHashFixture, Role: domainuser.RoleCustomer, TermsVersion: &version,
		})
		require.NoError(t, err)
		created, err := f.users.Create(ctx, port.CreateUserParams{
			Email: "created-by-admin@example.com", PasswordHash: testPasswordHashFixture, Role: domainuser.RoleStaff,
		})
		require.NoError(t, err)

		accepted, err := f.users.TermsAcceptance(ctx, signedUp.ID)
		require.NoError(t, err)
		require.NotNil(t, accepted)
		assert.Equal(t, domainpolicy.TermsVersion, accepted.Version)
		assert.WithinDuration(t, signedUp.CreatedAt, accepted.AcceptedAt, 0, "accepted as the account was created")

		none, err := f.users.TermsAcceptance(ctx, created.ID)
		require.NoError(t, err)
		assert.Nil(t, none, "an account nobody signed up for accepted nothing")
	})

	t.Run("an_order_records_the_version_accepted_at_checkout", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		placed, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(9, 0)))
		require.NoError(t, err)

		order, err := f.orders.GetByIDForUser(ctx, customer, uuid.MustParse(placed.ID))
		require.NoError(t, err)
		require.NotNil(t, order.Terms)
		assert.Equal(t, domainpolicy.TermsVersion, order.Terms.Version)
		assert.WithinDuration(t, order.CreatedAt, order.Terms.AcceptedAt, 0, "accepted as the order was placed")
	})

	t.Run("a_customer_exports_their_account_cart_sessions_activity_and_orders", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.cake, f.pastry}, nil)
		name, phone := "Mai", "+84901234567"
		_, err := customerProfiles.Create(ctx, port.UpsertCustomerProfileParams{UserID: customer, DisplayName: &name, Phone: &phone})
		require.NoError(t, err)
		first, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(10, 0)))
		require.NoError(t, err)
		require.NoError(t, f.fillCart(customer, []uuid.UUID{f.pastry}, []uuid.UUID{f.combo}))
		second, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(11, 0)))
		require.NoError(t, err)
		require.NoError(t, f.fillCart(customer, []uuid.UUID{f.cake}, nil))
		f.newCustomer(t, []uuid.UUID{f.pastry}, nil)

		admin := f.newWorker(t, domainuser.RoleAdmin)
		ownIP, adminIP := "198.51.100.4", "192.0.2.9"
		require.NoError(t, audit.Create(ctx, port.CreateAuditLogParams{
			ActorID: customer, ActorRole: domainuser.RoleCustomer, Action: domainuser.AuditActionMeUpdatedProfile,
			TargetID: &customer, TargetType: "user_profile", BeforeJSON: []byte(`{"phone":null}`), AfterJSON: []byte(`{"phone":"+84901234567"}`),
			IP: &ownIP,
		}))
		require.NoError(t, audit.Create(ctx, port.CreateAuditLogParams{
			ActorID: admin, ActorRole: domainuser.RoleAdmin, Action: domainuser.AuditActionAdminEnabledUser,
			TargetID: &customer, TargetType: "user", BeforeJSON: []byte(`{"disabled":true}`), AfterJSON: []byte(`{"disabled":false}`),
			IP: &adminIP,
		}))

		export, err := exports.Export(ctx, customer)

		require.NoError(t, err)
		assert.Equal(t, customer, export.User.ID)
		assert.NotNil(t, export.Profile, "the profile they gave")
		require.Len(t, export.Orders, 2, "their orders and nobody else's")
		assert.Equal(t, []string{second.ID, first.ID}, []string{export.Orders[0].ID, export.Orders[1].ID}, "newest first")
		for _, order := range export.Orders {
			assert.Len(t, order.Items, 2)
			assert.Equal(t, "awaiting_payment", order.Timeline[0].Status)
			require.NotNil(t, order.TermsAcceptance)
			assert.Equal(t, domainpolicy.TermsVersion, order.TermsAcceptance.Version)
		}
		require.Len(t, export.Cart, 1, "the cake still waiting in the cart")
		assert.Equal(t, f.cake.String(), *export.Cart[0].ProductID)
		require.Len(t, export.Sessions, 1)
		assert.Equal(t, dto.DataExportSessionResponse{
			SignedInAt: sessions[0].CreatedAt, IP: "203.0.113.7", UserAgent: "Test browser",
		}, export.Sessions[0])
		written, err := json.Marshal(export.Sessions)
		require.NoError(t, err)
		assert.NotContains(t, string(written), "refresh-id-that-stays-private", "a session's secrets never leave it")

		require.Len(t, export.Activity, 2)
		byAdmin, own := export.Activity[0], export.Activity[1]
		assert.Equal(t, dto.AccountActivityByYou, own.By)
		require.NotNil(t, own.IP)
		assert.Equal(t, ownIP, *own.IP, "their own network, for their own change")
		assert.JSONEq(t, `{"phone":"+84901234567"}`, string(own.After))
		assert.Equal(t, string(domainuser.RoleAdmin), byAdmin.By)
		assert.Nil(t, byAdmin.IP, "a staff member's network is theirs, not the customer's")
	})

	t.Run("an_export_reads_every_order_across_pages_exactly_once", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		const placed = 101
		for i := 1; i <= placed; i++ {
			_, err := f.orders.Create(ctx, port.CreateOrderParams{
				UserID: &customer, Channel: domainorder.ChannelOnline, Code: fmt.Sprintf("CH-250101-%03d", i), Status: domainorder.StatusFulfilled,
				Type: domainorder.TypeInstant, SubtotalCents: 300, TotalCents: 300,
			})
			require.NoError(t, err)
		}

		export, err := exports.Export(ctx, customer)

		require.NoError(t, err)
		require.Len(t, export.Orders, placed, "one page of 100 and one of 1")
		codes := make([]string, 0, placed)
		for i, order := range export.Orders {
			codes = append(codes, order.Code)
			if i > 0 {
				assert.False(t, order.CreatedAt.After(export.Orders[i-1].CreatedAt), "newest first")
			}
		}
		slices.Sort(codes)
		assert.Len(t, slices.Compact(codes), placed, "no order twice")
	})

	t.Run("a_worker_exports_their_account_and_no_orders", func(t *testing.T) {
		worker := f.newWorker(t, domainuser.RoleStaff)

		export, err := exports.Export(ctx, worker)

		require.NoError(t, err)
		assert.Equal(t, domainuser.RoleStaff, export.User.Role)
		assert.Nil(t, export.Profile, "an account without a profile exports none")
		assert.Nil(t, export.Terms)
		assert.Empty(t, export.Orders)
		assert.Empty(t, export.Cart)
		assert.Empty(t, export.Activity)
	})
}
