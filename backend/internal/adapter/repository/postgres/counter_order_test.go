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
	domaincart "github.com/boms/backend/internal/domain/cart"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// The counter takes orders for walk-in guests and, on the phone, for
// customers with an account: priced and booked as a checkout is, confirmed at
// once, and paid in cash when collected. A product that runs out stays out for
// the rest of its day.
func TestCounterOrders_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	products := postgresadapter.NewProductRepository(f.pool)
	payments := postgresadapter.NewPaymentRepository(f.pool)
	soldOut := usecase.NewStaffProductUsecase(products, f.pool, f.outbox, nil, zap.NewNop())

	take := func(key uuid.UUID, req dto.CreateStaffOrderRequest) (*dto.StaffOrderResponse, error) {
		return f.staff.CreateOrder(ctx, f.clerk, domainuser.RoleStaff, key, req)
	}
	pastry := f.pastry.String()
	guestOrder := func(hour int) dto.CreateStaffOrderRequest {
		return dto.CreateStaffOrderRequest{
			Channel:  "counter",
			Guest:    &dto.StaffOrderGuestRequest{Name: " Lan ", Phone: "0901234567"},
			PickupAt: tomorrowAt(hour, 0),
			Items:    []dto.StaffOrderItemRequest{{ProductID: &pastry, Quantity: 2}},
		}
	}
	validationCode := func(t *testing.T, err error) string {
		t.Helper()
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr)
		return appErr.Code
	}

	t.Run("a_guest_order_is_confirmed_with_its_cash_due", func(t *testing.T) {
		order, err := take(uuid.New(), guestOrder(9))

		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusConfirmed), order.Status)
		assert.Equal(t, string(domainorder.ChannelCounter), order.Channel)
		assert.Equal(t, int64(600), order.TotalCents, "priced from the catalog")
		assert.Nil(t, order.Customer.UserID)
		assert.Nil(t, order.Customer.Email)
		assert.Equal(t, "Lan", *order.Customer.DisplayName)
		assert.Equal(t, "+84901234567", *order.Customer.Phone)
		require.NotNil(t, order.Payment)
		assert.Equal(t, string(domainpayment.ProviderCash), order.Payment.Provider)
		assert.Equal(t, string(domainpayment.StatusCreated), order.Payment.Status)
		assert.Len(t, order.Tickets, 1, "the counter makes it")

		list, _, _, _, err := f.staff.List(ctx, 1, 100, "")
		require.NoError(t, err)
		listed := findSummary(t, list, order.ID)
		assert.Equal(t, "Lan", *listed.Customer.DisplayName)
		assert.Nil(t, listed.Customer.Phone, "a list never hands out phones")
	})

	t.Run("a_quote_prices_the_items_and_says_what_they_ask", func(t *testing.T) {
		cake := f.cake.String()
		quote, err := f.staff.Quote(ctx, dto.StaffOrderQuoteRequest{Items: []dto.StaffOrderItemRequest{
			{ProductID: &pastry, Quantity: 2}, {ProductID: &cake, Quantity: 1},
		}})

		require.NoError(t, err)
		assert.Equal(t, int64(2*300+4500), quote.TotalCents)
		assert.True(t, quote.Fulfillment.HasKitchenItems)
		assert.Equal(t, 180, quote.Fulfillment.LeadMinutes)
	})

	t.Run("the_same_key_takes_the_order_once", func(t *testing.T) {
		key := uuid.New()
		first, err := take(key, guestOrder(9))
		require.NoError(t, err)
		again, err := take(key, guestOrder(9))
		require.NoError(t, err)
		assert.Equal(t, first.ID, again.ID)
	})

	t.Run("a_retry_racing_the_first_request_gets_its_order", func(t *testing.T) {
		key := uuid.New()
		ids := make(chan string, 2)
		errs := make(chan error, 2)
		for range 2 {
			go func() {
				order, err := take(key, guestOrder(15))
				errs <- err
				if err == nil {
					ids <- order.ID
				}
			}()
		}
		require.NoError(t, <-errs)
		require.NoError(t, <-errs)
		assert.Equal(t, <-ids, <-ids)
	})

	t.Run("the_counter_finds_a_customer_account_by_its_email", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		account, err := f.users.GetByID(ctx, customer)
		require.NoError(t, err)

		found, err := f.staff.FindCustomer(ctx, " "+account.Email+" ")
		require.NoError(t, err)
		assert.Equal(t, customer.String(), found.ID)

		_, err = f.staff.FindCustomer(ctx, "nobody@example.com")
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
		clerk, err := f.users.GetByID(ctx, f.clerk)
		require.NoError(t, err)
		_, err = f.staff.FindCustomer(ctx, clerk.Email)
		assert.ErrorIs(t, err, apperrors.ErrNotFound, "staff accounts are not customers")
	})

	t.Run("a_phone_order_is_its_customers_to_see", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		id := customer.String()
		req := guestOrder(10)
		req.Channel, req.Guest, req.CustomerID = "phone", nil, &id

		order, err := take(uuid.New(), req)

		require.NoError(t, err)
		assert.Equal(t, &id, order.Customer.UserID)
		mine, err := f.orderUC.Get(ctx, customer, uuid.MustParse(order.ID))
		require.NoError(t, err)
		require.NotNil(t, mine.Payment)
		assert.Equal(t, string(domainpayment.ProviderCash), mine.Payment.Provider)
		assert.Nil(t, mine.PickupCode, "it is handed over as its cash is paid")
	})

	t.Run("an_order_is_for_one_account_or_one_guest_and_sold_items", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil).String()
		both := guestOrder(11)
		both.CustomerID = &customer
		_, err := take(uuid.New(), both)
		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, err), "both")

		neither := guestOrder(11)
		neither.Guest = nil
		_, err = take(uuid.New(), neither)
		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, err), "neither")

		nobody := uuid.New().String()
		unknown := guestOrder(11)
		unknown.Guest, unknown.CustomerID = nil, &nobody
		_, err = take(uuid.New(), unknown)
		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, err), "no such account")

		clerk := f.clerk.String()
		staffAccount := guestOrder(11)
		staffAccount.Guest, staffAccount.CustomerID = nil, &clerk
		_, err = take(uuid.New(), staffAccount)
		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, err), "not a customer account")

		landline := guestOrder(11)
		landline.Guest.Phone = "028 3822 1234"
		_, err = take(uuid.New(), landline)
		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, err), "not a mobile number")

		online := guestOrder(11)
		online.Channel = "online"
		_, err = take(uuid.New(), online)
		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, err), "staff do not take online orders")

		unnamed := guestOrder(11)
		unnamed.Guest.Name = "  "
		_, err = take(uuid.New(), unnamed)
		assert.ErrorIs(t, err, domainorder.ErrInvalidGuestName)

		twice := guestOrder(11)
		twice.Items = append(twice.Items, twice.Items[0])
		_, err = take(uuid.New(), twice)
		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, err), "a product twice")

		retired, err := products.Create(ctx, port.CreateProductParams{
			CategoryID: categoryOf(t, raw, f.pastry), Name: "Old bun", Slug: "old-bun", PriceCents: 200, IsActive: false,
		})
		require.NoError(t, err)
		gone := retired.ID.String()
		offMenu := guestOrder(11)
		offMenu.Items = []dto.StaffOrderItemRequest{{ProductID: &gone, Quantity: 1}}
		_, err = take(uuid.New(), offMenu)
		assert.ErrorIs(t, err, domaincart.ErrProductUnavailable)
	})

	t.Run("cash_taken_at_the_counter_hands_the_order_over", func(t *testing.T) {
		order, err := take(uuid.New(), guestOrder(12))
		require.NoError(t, err)
		_, err = raw.Exec(ctx, "UPDATE orders SET status = 'ready' WHERE id = $1", order.ID)
		require.NoError(t, err)
		handOver := func(req dto.PatchStaffOrderStatusRequest) error {
			_, err := f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID), req)
			return err
		}

		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, handOver(dto.PatchStaffOrderStatusRequest{Status: "fulfilled"})),
			"the cash first")
		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, handOver(dto.PatchStaffOrderStatusRequest{
			Status: "fulfilled", PickupCode: "1234", CashCollected: true,
		})), "no code to give")
		require.NoError(t, handOver(dto.PatchStaffOrderStatusRequest{Status: "fulfilled", CashCollected: true}))

		paid, err := payments.GetByOrderID(ctx, uuid.MustParse(order.ID))
		require.NoError(t, err)
		assert.Equal(t, domainpayment.StatusCaptured, paid.Status)
		assert.NotNil(t, paid.CapturedAt)
	})

	t.Run("an_order_with_nothing_to_pay_is_handed_over_without_cash", func(t *testing.T) {
		tasting, err := products.Create(ctx, port.CreateProductParams{
			CategoryID: categoryOf(t, raw, f.pastry), Name: "Tasting bite", Slug: "tasting-bite", PriceCents: 0, IsActive: true,
		})
		require.NoError(t, err)
		free := tasting.ID.String()
		req := guestOrder(16)
		req.Items = []dto.StaffOrderItemRequest{{ProductID: &free, Quantity: 1}}
		order, err := take(uuid.New(), req)
		require.NoError(t, err)
		assert.Nil(t, order.Payment, "no cash due")
		_, err = raw.Exec(ctx, "UPDATE orders SET status = 'ready' WHERE id = $1", order.ID)
		require.NoError(t, err)

		handed, err := f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.PatchStaffOrderStatusRequest{Status: "fulfilled", CashCollected: true})

		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusFulfilled), handed.Status)
	})

	t.Run("an_online_order_still_takes_its_code_not_cash", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 13)
		_, err := raw.Exec(ctx, "UPDATE orders SET status = 'ready' WHERE id = $1", order.ID)
		require.NoError(t, err)

		_, err = f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.PatchStaffOrderStatusRequest{Status: "fulfilled", CashCollected: true})

		assert.Equal(t, apperrors.ErrValidation.Code, validationCode(t, err))
	})

	// The day an item is out comes back with a cart that holds it, so its picker
	// and its checkout refuse that day (the rule itself is the pickup contract's).
	t.Run("the_counter_marks_a_product_sold_out_for_today", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		marked, err := soldOut.SetSoldOut(ctx, f.clerk, domainuser.RoleStaff, f.pastry, true)
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = soldOut.SetSoldOut(context.Background(), f.clerk, domainuser.RoleStaff, f.pastry, false)
		})

		assert.True(t, marked.SoldOutToday)
		out, total, _, _, err := soldOut.List(ctx, 1, 100, true)
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, f.pastry.String(), out[0].ID)
		cart, err := f.cartUC.Get(ctx, customer)
		require.NoError(t, err)
		today := domainstore.DayOf(time.Now()).Format(domainstore.DayLayout)
		assert.Equal(t, &today, cart.Fulfillment.SoldOutOn)

		back, err := soldOut.SetSoldOut(ctx, f.clerk, domainuser.RoleStaff, f.pastry, false)
		require.NoError(t, err)
		assert.False(t, back.SoldOutToday)
		cart, err = f.cartUC.Get(ctx, customer)
		require.NoError(t, err)
		assert.Nil(t, cart.Fulfillment.SoldOutOn)
		_, err = soldOut.SetSoldOut(ctx, f.clerk, domainuser.RoleStaff, uuid.New(), true)
		assert.ErrorIs(t, err, domainproduct.ErrNotFound, "no such product")
	})
}

func findSummary(t *testing.T, list []dto.StaffOrderSummaryResponse, id string) dto.StaffOrderSummaryResponse {
	t.Helper()
	for _, order := range list {
		if order.ID == id {
			return order
		}
	}
	require.FailNow(t, "order not listed", id)
	return dto.StaffOrderSummaryResponse{}
}

func categoryOf(t *testing.T, raw *pgxpool.Pool, productID uuid.UUID) uuid.UUID {
	t.Helper()
	var category uuid.UUID
	require.NoError(t, raw.QueryRow(context.Background(), "SELECT category_id FROM products WHERE id = $1", productID).Scan(&category))
	return category
}
