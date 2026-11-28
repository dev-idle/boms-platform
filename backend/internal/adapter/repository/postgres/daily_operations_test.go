package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/usecase"
)

// The day's loose ends: an order missed and collected late, a pickup staff
// move, the orders left open when a day closed, a handover without the code
// that a manager signs for, and a day the bakery closes with orders on it.
func TestDailyOperations_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	setStatus := func(orderID string, status domainorder.Status, pickupAt time.Time) {
		_, err := raw.Exec(ctx, "UPDATE orders SET status = $2, pickup_at = $3 WHERE id = $1", orderID, string(status), pickupAt)
		require.NoError(t, err)
	}
	manager := f.newWorker(t, domainuser.RoleManager)
	audit := auditlogger.NewService(postgresadapter.NewAuditLogRepository(f.pool))
	discounts := postgresadapter.NewDiscountCodeRepository(f.pool)
	payments := postgresadapter.NewPaymentRepository(f.pool)
	managerUC := usecase.NewManagerOrderUsecase(f.orders, f.tickets, discounts, payments, f.pool, f.outbox, audit)
	settingsUC := usecase.NewManagerStoreSettingsUsecase(f.store, f.orders, f.tickets, discounts, payments, f.paymentUC, f.pool, f.outbox, audit, zap.NewNop())
	audited := func(action, orderID string) int {
		var rows int
		require.NoError(t, raw.QueryRow(ctx, "SELECT count(*) FROM audit_logs WHERE action = $1 AND target_id = $2", action, orderID).Scan(&rows))
		return rows
	}
	statusOf := func(orderID string) string {
		got, err := f.staff.Get(ctx, uuid.MustParse(orderID))
		require.NoError(t, err)
		return got.Status
	}

	t.Run("a_missed_order_is_still_handed_over_with_its_code", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 9)
		setStatus(order.ID, domainorder.StatusNoShow, time.Now().Add(-48*time.Hour))

		handed, err := f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID), dto.PatchStaffOrderStatusRequest{
			Status: "fulfilled", PickupCode: testPickupCodes.Of(uuid.MustParse(order.ID)),
		})

		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusFulfilled), handed.Status)
	})

	t.Run("staff_move_a_ready_order_to_a_later_pickup", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 9)
		setStatus(order.ID, domainorder.StatusReady, tomorrowAt(9, 0))

		moved, err := f.staff.MovePickup(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.RescheduleOrderRequest{PickupAt: tomorrowAt(17, 30)})

		require.NoError(t, err)
		require.NotNil(t, moved.PickupAt)
		assert.True(t, moved.PickupAt.Equal(tomorrowAt(17, 30)))
		assert.Equal(t, string(domainorder.StatusReady), moved.Status, "the order stays as made")
		_, err = f.staff.MovePickup(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.RescheduleOrderRequest{PickupAt: tomorrowAt(17, 45)})
		require.ErrorIs(t, err, domainorder.ErrPickupOffSlot)
	})

	t.Run("a_move_to_a_full_slot_or_the_same_time_changes_nothing", func(t *testing.T) {
		f.setSettings(t, func(s *domainstore.Settings) { s.SlotCapacity = 1 })
		t.Cleanup(func() { f.setSettings(t, func(s *domainstore.Settings) { s.SlotCapacity = 8 }) })
		taken := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 15)
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 16)

		_, err := f.staff.MovePickup(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.RescheduleOrderRequest{PickupAt: tomorrowAt(15, 0)})
		require.ErrorIs(t, err, domainorder.ErrPickupSlotFull, "15:00 holds %s already", taken.Code)

		same, err := f.staff.MovePickup(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.RescheduleOrderRequest{PickupAt: tomorrowAt(16, 0)})
		require.NoError(t, err)
		assert.True(t, same.PickupAt.Equal(tomorrowAt(16, 0)))
		assert.Zero(t, audited("staff.moved_pickup", order.ID), "nothing moved, nothing audited")
	})

	t.Run("an_order_handed_over_is_not_moved", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 10)
		setStatus(order.ID, domainorder.StatusFulfilled, tomorrowAt(10, 0))

		_, err := f.staff.MovePickup(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.RescheduleOrderRequest{PickupAt: tomorrowAt(11, 0)})

		require.ErrorIs(t, err, domainorder.ErrInvalidStatusTransition)
	})

	t.Run("an_order_left_open_when_its_day_closed_is_flagged_once_and_listed", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 11)
		setStatus(order.ID, domainorder.StatusInProduction, time.Now().Add(-48*time.Hour))
		later := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 11)

		flagged, err := f.orderUC.FlagLeftOpen(ctx)

		require.NoError(t, err)
		assert.Equal(t, 1, flagged)
		var incidents int
		require.NoError(t, raw.QueryRow(ctx, "SELECT count(*) FROM order_incidents WHERE order_id = $1 AND type = 'left_open'", order.ID).Scan(&incidents))
		assert.Equal(t, 1, incidents)
		again, err := f.orderUC.FlagLeftOpen(ctx)
		require.NoError(t, err)
		assert.Zero(t, again, "an order is flagged once")
		listed, total, _, _, err := f.staff.List(ctx, 1, 100, usecase.StaffOrderFilterLeftOpen, "")
		require.NoError(t, err)
		ids := make([]string, 0, len(listed))
		for _, row := range listed {
			ids = append(ids, row.ID)
		}
		assert.Contains(t, ids, order.ID)
		assert.NotContains(t, ids, later.ID, "its pickup is still ahead")
		assert.Equal(t, int64(len(listed)), total)
	})

	t.Run("a_manager_hands_over_an_online_order_without_its_code", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 12)
		setStatus(order.ID, domainorder.StatusReady, tomorrowAt(12, 0))

		found, err := managerUC.Lookup(ctx, " "+order.Code+" ")
		require.NoError(t, err)
		assert.Equal(t, order.ID, found.ID)
		_, err = managerUC.HandOver(ctx, manager, domainuser.RoleManager, uuid.MustParse(order.ID), dto.ManagerHandoverRequest{Reason: "  "})
		require.ErrorIs(t, err, domainorder.ErrInvalidHandoverReason)
		handed, err := managerUC.HandOver(ctx, manager, domainuser.RoleManager, uuid.MustParse(order.ID),
			dto.ManagerHandoverRequest{Reason: "Checked the buyer's ID and PayPal receipt"})

		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusFulfilled), handed.Status)
		assert.Equal(t, 1, audited("manager.handed_over_without_code", order.ID), "the reason commits with the handover")
		_, err = managerUC.HandOver(ctx, manager, domainuser.RoleManager, uuid.MustParse(order.ID),
			dto.ManagerHandoverRequest{Reason: "Again"})
		require.ErrorIs(t, err, domainorder.ErrInvalidStatusTransition)
	})

	t.Run("new_hours_find_the_pickups_they_leave_outside", func(t *testing.T) {
		early := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 8)
		setStatus(early.ID, domainorder.StatusConfirmed, tomorrowAt(8, 0).AddDate(0, 0, 1))
		midday := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 12)
		setStatus(midday.ID, domainorder.StatusConfirmed, tomorrowAt(12, 0).AddDate(0, 0, 1))

		stranded, err := settingsUC.StrandedPickups(ctx, dto.StrandedPickupsQuery{OpensAt: "09:00", ClosesAt: "18:00"})

		require.NoError(t, err)
		codes := make([]string, 0, len(stranded.Orders))
		for _, o := range stranded.Orders {
			codes = append(codes, o.Code)
		}
		assert.Contains(t, codes, early.Code, "08:00 is before the new opening")
		assert.NotContains(t, codes, midday.Code)
		opens := "09:00"
		_, err = settingsUC.PatchSettings(ctx, manager, domainuser.RoleManager, dto.PatchStoreSettingsRequest{OpensAt: &opens})
		require.ErrorIs(t, err, domainstore.ErrPickupsStranded)
	})

	t.Run("the_manager_hands_over_only_a_ready_or_missed_online_order", func(t *testing.T) {
		pastry := f.pastry.String()
		counter, err := f.staff.CreateOrder(ctx, f.clerk, domainuser.RoleStaff, uuid.New(), dto.CreateStaffOrderRequest{
			Channel:  "counter",
			Guest:    &dto.StaffOrderGuestRequest{Name: "Lan", Phone: "0901234567"},
			PickupAt: tomorrowAt(12, 30),
			Items:    []dto.StaffOrderItemRequest{{ProductID: &pastry, Quantity: 1}},
		})
		require.NoError(t, err)
		setStatus(counter.ID, domainorder.StatusReady, tomorrowAt(12, 30))
		confirmed := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 12)
		unpaid, err := f.orderUC.Checkout(ctx, f.newCustomer(t, []uuid.UUID{f.pastry}, nil), uuid.New(), acceptingTerms(tomorrowAt(12, 30)))
		require.NoError(t, err)
		handover := dto.ManagerHandoverRequest{Reason: "Checked the buyer's ID"}

		_, err = managerUC.HandOver(ctx, manager, domainuser.RoleManager, uuid.MustParse(counter.ID), handover)
		require.ErrorIs(t, err, domainorder.ErrInvalidStatusTransition, "the counter hands its own order over with its cash")
		_, err = managerUC.HandOver(ctx, manager, domainuser.RoleManager, uuid.MustParse(confirmed.ID), handover)
		require.ErrorIs(t, err, domainorder.ErrInvalidStatusTransition, "not ready yet")
		_, err = managerUC.Lookup(ctx, unpaid.Code)
		require.ErrorIs(t, err, domainorder.ErrNotFound, "an unpaid order is the customer's alone")
	})

	t.Run("an_order_under_review_at_paypal_keeps_the_day_open", func(t *testing.T) {
		later := tomorrowAt(14, 0).AddDate(0, 0, 1)
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		held, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(later))
		require.NoError(t, err)
		_, err = f.paymentUC.Start(ctx, customer, uuid.MustParse(held.ID))
		require.NoError(t, err)
		f.paypal.mu.Lock()
		f.paypal.taken["PAYPAL-"+held.ID] = domainpayment.StatusPending
		f.paypal.mu.Unlock()
		day := domainstore.DayOf(later).Format(domainstore.DayLayout)

		_, err = settingsUC.CancelStrandedPickups(ctx, manager, domainuser.RoleManager, dto.CancelStrandedPickupsRequest{
			StrandedPickupsQuery: dto.StrandedPickupsQuery{ClosedDate: day}, Reason: "Staff training",
		})

		require.ErrorIs(t, err, domainpayment.ErrUnderReview)
		_, err = settingsUC.AddClosedDate(ctx, manager, domainuser.RoleManager, dto.CreateClosedDateRequest{Date: day, Reason: "Staff training"})
		require.ErrorIs(t, err, domainstore.ErrPickupsStranded, "PayPal may still take the money for that day")
	})

	t.Run("a_day_closes_once_its_orders_are_cancelled_and_refunded", func(t *testing.T) {
		paid := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 13)
		unpaidCustomer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		unpaid, err := f.orderUC.Checkout(ctx, unpaidCustomer, uuid.New(), acceptingTerms(tomorrowAt(13, 30)))
		require.NoError(t, err)
		day := domainstore.DayOf(tomorrowAt(12, 0)).Format(domainstore.DayLayout)
		closing := dto.CreateClosedDateRequest{Date: day, Reason: "Power cut"}

		_, err = settingsUC.AddClosedDate(ctx, manager, domainuser.RoleManager, closing)
		require.ErrorIs(t, err, domainstore.ErrPickupsStranded)
		stranded, err := settingsUC.StrandedPickups(ctx, dto.StrandedPickupsQuery{ClosedDate: day})
		require.NoError(t, err)
		codes := make([]string, 0, len(stranded.Orders))
		for _, o := range stranded.Orders {
			codes = append(codes, o.Code)
		}
		assert.Contains(t, codes, paid.Code)
		assert.Contains(t, codes, unpaid.Code, "an unpaid order would be paid for a closed day")

		done, err := settingsUC.CancelStrandedPickups(ctx, manager, domainuser.RoleManager, dto.CancelStrandedPickupsRequest{
			StrandedPickupsQuery: dto.StrandedPickupsQuery{ClosedDate: day}, Reason: "We have a power cut that day.",
		})

		require.NoError(t, err)
		assert.Equal(t, int(stranded.Total), done.Cancelled)
		assert.Equal(t, string(domainorder.StatusCancelled), statusOf(paid.ID))
		after, err := f.staff.Get(ctx, uuid.MustParse(paid.ID))
		require.NoError(t, err)
		require.NotNil(t, after.Payment.RefundRequestedAt, "a paid order is refunded")
		last := after.Timeline[len(after.Timeline)-1]
		require.NotNil(t, last.Reason)
		assert.Equal(t, "We have a power cut that day.", *last.Reason)
		assert.Equal(t, 1, audited("manager.cancelled_stranded_order", paid.ID), "each cancellation commits with its audit row")
		_, err = settingsUC.AddClosedDate(ctx, manager, domainuser.RoleManager, closing)
		require.NoError(t, err, "nothing is stranded any more")
		require.NoError(t, f.pool.WithTx(ctx, func(txCtx context.Context) error {
			held, err := f.store.HoldPickupRules(txCtx, domainstore.DayOf(tomorrowAt(12, 0)))
			require.NoError(t, err)
			assert.True(t, held.DayClosed, "a booking in flight finds the day closed under the held rules")
			return nil
		}))
	})
}
