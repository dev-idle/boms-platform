package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

const erasurePassword = "the-current-password"

// erasureCarts holds one cart, or none; any call an erasure should not make panics.
type erasureCarts struct {
	port.CartRepository
	cart    *domaincart.Cart
	emptied bool
}

func (c *erasureCarts) GetByUserIDForUpdate(context.Context, uuid.UUID) (*domaincart.Cart, error) {
	if c.cart == nil {
		return nil, apperrors.ErrNotFound
	}
	return c.cart, nil
}

func (c *erasureCarts) DeleteAllItems(context.Context, uuid.UUID) error {
	c.emptied = true
	return nil
}

func (c *erasureCarts) ClearDiscountCode(context.Context, uuid.UUID) error { return nil }

type erasureOrders struct {
	port.OrderRepository
	open bool
}

func (o *erasureOrders) HasOpen(context.Context, uuid.UUID) (bool, error) { return o.open, nil }

type erasureFixture struct {
	customer  *domainuser.User
	users     *mockUserRepo
	customers *mockCustomerProfileRepo
	carts     *erasureCarts
	orders    *erasureOrders
	audit     *recordingAuditLogs
	sessions  *mockSessionStore
	hasher    *mockHasher
}

// newErasureFixture is a signed-in customer with a cart, the right password
// and nothing open; each test changes what it is about.
func newErasureFixture() *erasureFixture {
	customer := &domainuser.User{ID: uuid.New(), Role: domainuser.RoleCustomer, PasswordHash: "stored-hash"}
	f := &erasureFixture{
		customer:  customer,
		users:     new(mockUserRepo),
		customers: new(mockCustomerProfileRepo),
		carts:     &erasureCarts{cart: &domaincart.Cart{ID: uuid.New(), UserID: customer.ID}},
		orders:    &erasureOrders{},
		audit:     &recordingAuditLogs{},
		sessions:  new(mockSessionStore),
		hasher:    new(mockHasher),
	}
	f.users.On("GetByID", mock.Anything, customer.ID).Return(customer, nil).Maybe()
	f.users.On("GetByIDForUpdate", mock.Anything, customer.ID).Return(customer, nil).Maybe()
	f.hasher.On("Verify", customer.PasswordHash, erasurePassword).Return(nil).Maybe()
	f.hasher.On("Verify", customer.PasswordHash, mock.Anything).Return(errors.New("mismatch")).Maybe()
	return f
}

func (f *erasureFixture) usecase() *usecase.AccountErasureUsecase {
	return usecase.NewAccountErasureUsecase(
		passthroughTxManager{}, f.users, f.customers, f.carts, f.orders, f.audit, f.sessions,
		auditlogger.NewService(f.audit), f.hasher,
	)
}

func (f *erasureFixture) expectErased() {
	f.customers.On("Erase", mock.Anything, f.customer.ID).Return(nil)
	f.users.On("Erase", mock.Anything, f.customer.ID).Return(nil)
}

// nothingErased checks a refusal left the account as it was.
func (f *erasureFixture) nothingErased(t *testing.T) {
	t.Helper()
	assert.False(t, f.carts.emptied, "the cart is kept")
	f.customers.AssertNotCalled(t, "Erase", mock.Anything, mock.Anything)
	f.users.AssertNotCalled(t, "Erase", mock.Anything, mock.Anything)
	f.sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
}

func TestAccountErasureUsecase_Erase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("erases_the_account_then_ends_its_sessions", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.expectErased()
		f.sessions.On("DeleteAllForUser", mock.Anything, f.customer.ID.String()).Return(nil)

		require.NoError(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword))

		assert.True(t, f.carts.emptied, "the cart is emptied")
		assert.Equal(t, []domainuser.AuditAction{domainuser.AuditActionMeErasedAccount}, f.audit.actions)
		assert.True(t, f.audit.scrubbed, "the audit trail loses the personal data")
		f.users.AssertCalled(t, "GetByIDForUpdate", mock.Anything, f.customer.ID)
		f.customers.AssertExpectations(t)
		f.users.AssertExpectations(t)
		f.sessions.AssertExpectations(t)
	})

	t.Run("a_customer_without_a_cart_is_erased_too", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.carts.cart = nil
		f.expectErased()
		f.sessions.On("DeleteAllForUser", mock.Anything, f.customer.ID.String()).Return(nil)

		require.NoError(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword))

		assert.False(t, f.carts.emptied, "there was no cart to empty")
		f.users.AssertExpectations(t)
	})

	t.Run("asks_for_the_right_password", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()

		require.ErrorIs(t, f.usecase().Erase(ctx, f.customer.ID, "a-guess"), apperrors.ErrInvalidCredentials)

		f.nothingErased(t)
		f.users.AssertNotCalled(t, "GetByIDForUpdate", mock.Anything, mock.Anything)
	})

	t.Run("staff_are_closed_by_an_administrator_not_by_themselves", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.customer.Role = domainuser.RoleStaff

		require.ErrorIs(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword), domainuser.ErrSelfDeleteCustomerOnly)

		f.nothingErased(t)
		f.hasher.AssertNotCalled(t, "Verify", mock.Anything, mock.Anything)
	})

	t.Run("waits_until_no_order_is_open", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.orders.open = true

		require.ErrorIs(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword), domainuser.ErrAccountHasOpenOrders)

		f.nothingErased(t)
		assert.Empty(t, f.audit.actions)
	})

	t.Run("an_account_closed_while_waiting_for_its_lock_is_not_found", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.users.ExpectedCalls = nil
		f.users.On("GetByID", mock.Anything, f.customer.ID).Return(f.customer, nil)
		f.users.On("GetByIDForUpdate", mock.Anything, f.customer.ID).Return(nil, apperrors.ErrNotFound)

		require.ErrorIs(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword), usecase.ErrMeNotFound)

		f.nothingErased(t)
	})

	t.Run("a_failed_step_erases_nothing", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.customers.On("Erase", mock.Anything, f.customer.ID).Return(apperrors.ErrInternal)

		require.ErrorIs(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword), apperrors.ErrInternal)

		f.users.AssertNotCalled(t, "Erase", mock.Anything, mock.Anything)
		f.sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
	})

	t.Run("an_erasure_that_cannot_be_recorded_does_not_happen", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.expectErased()
		f.audit.failWith = apperrors.ErrInternal

		require.Error(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword))

		assert.False(t, f.audit.scrubbed, "the transaction stops at the failed record")
		f.sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
	})

	t.Run("sessions_left_by_a_failed_sign_out_end_on_the_retry", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.expectErased()
		redisDown := errors.New("redis unavailable")
		f.sessions.On("DeleteAllForUser", mock.Anything, f.customer.ID.String()).Return(redisDown).Once()
		uc := f.usecase()

		require.ErrorIs(t, uc.Erase(ctx, f.customer.ID, erasurePassword), redisDown)

		// The erasure committed: the account is closed and marked erased.
		erasedAt := time.Now()
		f.users.ExpectedCalls = nil
		f.users.On("GetByID", mock.Anything, f.customer.ID).Return(nil, apperrors.ErrNotFound)
		f.users.On("AdminGetByID", mock.Anything, f.customer.ID).Return(&domainuser.User{
			ID: f.customer.ID, Role: domainuser.RoleCustomer, DeletedAt: &erasedAt, ErasedAt: &erasedAt,
		}, nil)
		f.sessions.On("DeleteAllForUser", mock.Anything, f.customer.ID.String()).Return(nil).Once()

		// The password no longer matches anything, so the retry does not ask for it.
		require.NoError(t, uc.Erase(ctx, f.customer.ID, ""))

		f.sessions.AssertExpectations(t)
		f.customers.AssertNumberOfCalls(t, "Erase", 1)
	})

	t.Run("a_disabled_account_is_not_found", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		disabledAt := time.Now()
		f.users.ExpectedCalls = nil
		f.users.On("GetByID", mock.Anything, f.customer.ID).Return(nil, apperrors.ErrNotFound)
		f.users.On("AdminGetByID", mock.Anything, f.customer.ID).Return(&domainuser.User{
			ID: f.customer.ID, Role: domainuser.RoleCustomer, DeletedAt: &disabledAt,
		}, nil)

		require.ErrorIs(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword), usecase.ErrMeNotFound)

		f.sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
	})

	t.Run("an_unknown_account_is_not_found", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.users.ExpectedCalls = nil
		f.users.On("GetByID", mock.Anything, mock.Anything).Return(nil, apperrors.ErrNotFound)
		f.users.On("AdminGetByID", mock.Anything, mock.Anything).Return(nil, apperrors.ErrNotFound)

		require.ErrorIs(t, f.usecase().Erase(ctx, uuid.New(), erasurePassword), usecase.ErrMeNotFound)
	})

	t.Run("a_failed_lookup_of_a_closed_account_is_reported", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.users.ExpectedCalls = nil
		f.users.On("GetByID", mock.Anything, f.customer.ID).Return(nil, apperrors.ErrNotFound)
		f.users.On("AdminGetByID", mock.Anything, f.customer.ID).Return(nil, apperrors.ErrInternal)

		require.ErrorIs(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword), apperrors.ErrInternal)

		f.sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
	})

	t.Run("a_failed_lookup_is_reported", func(t *testing.T) {
		t.Parallel()
		f := newErasureFixture()
		f.users.ExpectedCalls = nil
		f.users.On("GetByID", mock.Anything, f.customer.ID).Return(nil, apperrors.ErrInternal)

		require.ErrorIs(t, f.usecase().Erase(ctx, f.customer.ID, erasurePassword), apperrors.ErrInternal)

		f.nothingErased(t)
	})
}
