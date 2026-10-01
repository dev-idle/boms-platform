package postgres_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// The customer is given a pickup code once the bakery accepts the order and
// hands it over at the counter, where a ready order is collected only with it;
// the counter sees each day's pickups by time.
func TestPickupHandoff_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	setStatus := func(orderID string, status domainorder.Status) {
		_, err := raw.Exec(ctx, "UPDATE orders SET status = $2 WHERE id = $1", orderID, string(status))
		require.NoError(t, err)
	}
	move := func(staff *usecase.StaffOrderUsecase, orderID string, req dto.PatchStaffOrderStatusRequest) error {
		_, err := staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(orderID), req)
		return err
	}
	handOver := func(orderID, code string) error {
		return move(f.staff, orderID, dto.PatchStaffOrderStatusRequest{Status: "fulfilled", PickupCode: code})
	}
	statusOf := func(orderID string) string {
		order, err := f.staff.Get(ctx, uuid.MustParse(orderID))
		require.NoError(t, err)
		return order.Status
	}

	t.Run("the_customer_sees_the_code_once_the_bakery_accepts_the_order", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(9, 0)))
		require.NoError(t, err)
		assert.Nil(t, order.PickupCode, "not before it is paid")

		f.pay(t, customer, order.ID)

		paid, err := f.orderUC.Get(ctx, customer, uuid.MustParse(order.ID))
		require.NoError(t, err)
		require.NotNil(t, paid.PickupCode)
		assert.Equal(t, testPickupCodes.Of(uuid.MustParse(order.ID)), *paid.PickupCode)
	})

	t.Run("a_ready_order_is_handed_over_only_with_its_code", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 10)
		setStatus(order.ID, domainorder.StatusReady)
		code := testPickupCodes.Of(uuid.MustParse(order.ID))

		require.ErrorIs(t, handOver(order.ID, wrongCode(code)), domainorder.ErrPickupCodeInvalid)
		require.NoError(t, handOver(order.ID, code))

		assert.Equal(t, string(domainorder.StatusFulfilled), statusOf(order.ID))
	})

	t.Run("only_a_handover_carries_a_code", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 10)
		setStatus(order.ID, domainorder.StatusReady)

		var appErr *apperrors.AppError
		require.ErrorAs(t, handOver(order.ID, ""), &appErr)
		assert.Equal(t, apperrors.ErrValidation.Code, appErr.Code, "a handover without the code")
		require.ErrorAs(t, move(f.staff, order.ID, dto.PatchStaffOrderStatusRequest{
			Status: "cancelled", Reason: "Out of matcha", PickupCode: "1234",
		}), &appErr)
		assert.Equal(t, apperrors.ErrValidation.Code, appErr.Code, "a cancellation with a code")
		assert.Equal(t, string(domainorder.StatusReady), statusOf(order.ID))
	})

	t.Run("too_many_wrong_codes_hold_the_order", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 11)
		setStatus(order.ID, domainorder.StatusReady)
		code := testPickupCodes.Of(uuid.MustParse(order.ID))
		for range handoffLimit.Max {
			require.ErrorIs(t, handOver(order.ID, wrongCode(code)), domainorder.ErrPickupCodeInvalid)
		}

		require.ErrorIs(t, handOver(order.ID, code), domainorder.ErrPickupCodeLocked, "even the right code waits")
	})

	t.Run("no_handover_while_tries_cannot_be_counted", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 12)
		setStatus(order.ID, domainorder.StatusReady)
		staff := usecase.NewStaffOrderUsecase(f.orders, f.tickets, f.pool, f.outbox, nil, zap.NewNop(),
			postgresadapter.NewPaymentRepository(f.pool), postgresadapter.NewDiscountCodeRepository(f.pool), f.store, f.cartUC,
			testPickupCodes, brokenQuota{}, handoffLimit)

		err := move(staff, order.ID, dto.PatchStaffOrderStatusRequest{
			Status: "fulfilled", PickupCode: testPickupCodes.Of(uuid.MustParse(order.ID)),
		})

		require.Error(t, err)
		assert.Equal(t, string(domainorder.StatusReady), statusOf(order.ID))
	})

	t.Run("the_counter_sees_a_days_pickups_by_time", func(t *testing.T) {
		later := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 15)
		sooner := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 13)
		cancelled := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 14)
		setStatus(cancelled.ID, domainorder.StatusCancelled)
		day := domainstore.DayOf(tomorrowAt(12, 0)).Format(domainstore.DayLayout)

		schedule, total, _, _, err := f.staff.Pickups(ctx, day, 1, 100)

		require.NoError(t, err)
		settings, err := f.store.GetSettings(ctx)
		require.NoError(t, err)
		assert.Equal(t, int(settings.SlotLength.Minutes()), schedule.SlotMinutes)
		assert.Equal(t, int64(len(schedule.Pickups)), total)
		ids := pickupIDs(schedule.Pickups)
		assert.Less(t, slices.Index(ids, sooner.ID), slices.Index(ids, later.ID), "by pickup time")
		assert.NotContains(t, ids, cancelled.ID, "a cancelled order is not picked up")

		first, total, _, _, err := f.staff.Pickups(ctx, day, 1, 1)
		require.NoError(t, err)
		second, _, _, _, err := f.staff.Pickups(ctx, day, 2, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(len(ids)), total, "a page counts the whole day")
		assert.Equal(t, ids[:2], append(pickupIDs(first.Pickups), pickupIDs(second.Pickups)...), "pages follow on")

		other, _, _, _, err := f.staff.Pickups(ctx, domainstore.DayOf(tomorrowAt(12, 0).AddDate(0, 0, 1)).Format(domainstore.DayLayout), 1, 100)
		require.NoError(t, err)
		assert.Empty(t, other.Pickups)
		_, _, _, _, err = f.staff.Pickups(ctx, "tomorrow", 1, 100)
		require.Error(t, err)
	})

	t.Run("the_order_list_leads_with_the_soonest_open_pickup", func(t *testing.T) {
		openLater := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 17)
		openSooner := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 8)
		closedSooner := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 9)
		closedLater := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 16)
		setStatus(closedSooner.ID, domainorder.StatusFulfilled)
		setStatus(closedLater.ID, domainorder.StatusNoShow)

		list, _, _, _, err := f.staff.List(ctx, 1, 100, "")

		require.NoError(t, err)
		ids := pickupIDs(list)
		order := []string{openSooner.ID, openLater.ID, closedLater.ID, closedSooner.ID}
		positions := make([]int, 0, len(order))
		for _, id := range order {
			require.Contains(t, ids, id)
			positions = append(positions, slices.Index(ids, id))
		}
		assert.True(t, slices.IsSorted(positions),
			"open orders soonest pickup first, then closed ones latest pickup first")
	})
}

// brokenQuota cannot count tries, as when Redis is down.
type brokenQuota struct{}

func (brokenQuota) Take(context.Context, string, port.QuotaLimit) (bool, error) {
	return false, errors.New("redis is down")
}

// wrongCode is any four digits but code.
func wrongCode(code string) string {
	if code == "0000" {
		return "0001"
	}
	return "0000"
}

func pickupIDs(orders []dto.StaffOrderSummaryResponse) []string {
	ids := make([]string, 0, len(orders))
	for _, order := range orders {
		ids = append(ids, order.ID)
	}
	return ids
}
