package postgres_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	"github.com/boms/backend/internal/config"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// checkoutFixture is a bakery with a kitchen cake, a counter pastry and a combo
// of both, and a way to fill a new customer's cart.
type checkoutFixture struct {
	pool      *postgresadapter.Pool
	connStr   string
	users     *postgresadapter.UserRepository
	carts     *postgresadapter.CartRepository
	orders    *postgresadapter.OrderRepository
	store     *postgresadapter.StoreSettingsRepository
	outbox    *postgresadapter.OutboxRepository
	cartUC    *usecase.CartUsecase
	orderUC   *usecase.OrderUsecase
	paymentUC *usecase.PaymentUsecase
	paypal    *paidPayPal
	cake      uuid.UUID
	pastry    uuid.UUID
	combo     uuid.UUID
	customer  int
}

// testPickupCodes signs pickup codes as the API does, with a test key.
var testPickupCodes = domainorder.NewPickupCodes(strings.Repeat("k", 32))

// handoffLimit is how many pickup code tries an order takes in the tests.
var handoffLimit = port.QuotaLimit{Max: 3, Window: 15 * time.Minute}

// handoffAttempts counts pickup code tries per order in memory, as the Redis
// quota does in a window.
type handoffAttempts struct {
	mu    sync.Mutex
	taken map[string]int
}

func newHandoffAttempts() *handoffAttempts {
	return &handoffAttempts{taken: map[string]int{}}
}

func (q *handoffAttempts) Take(_ context.Context, key string, limit port.QuotaLimit) (bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.taken[key]++
	return q.taken[key] <= limit.Max, nil
}

// testCloudinary lets reference photos be attached: a customer's own folder
// is boms/references/<their id> on the demo cloud.
var testCloudinary = config.CloudinaryConfig{CloudName: "demo", APIKey: "key", APISecret: "secret"}

func newCheckoutFixture(t *testing.T, maxConns int32) *checkoutFixture {
	t.Helper()
	ctx := context.Background()
	pool, connStr := newIntegrationDB(t, maxConns)
	f := &checkoutFixture{
		pool:    pool,
		connStr: connStr,
		users:   postgresadapter.NewUserRepository(pool),
		carts:   postgresadapter.NewCartRepository(pool),
		orders:  postgresadapter.NewOrderRepository(pool),
		store:   postgresadapter.NewStoreSettingsRepository(pool),
		outbox:  postgresadapter.NewOutboxRepository(pool),
	}
	categories := postgresadapter.NewCategoryRepository(pool)
	products := postgresadapter.NewProductRepository(pool)
	combos := postgresadapter.NewComboRepository(pool)
	discounts := postgresadapter.NewDiscountCodeRepository(pool)
	tickets := postgresadapter.NewTicketRepository(pool)
	payments := postgresadapter.NewPaymentRepository(pool)
	f.cartUC = usecase.NewCartUsecase(f.carts, products, combos, discounts, testCloudinary)
	f.paypal = &paidPayPal{amounts: map[string]int64{}, taken: map[string]domainpayment.Status{}}
	f.paymentUC = usecase.NewPaymentUsecase(pool, f.orders, discounts, tickets, payments, f.paypal,
		f.outbox, "https://shop.example", zap.NewNop())
	f.orderUC = usecase.NewOrderUsecase(f.users, f.orders, f.carts, discounts, f.cartUC, pool, f.outbox, f.store, tickets, payments, f.paymentUC, testPickupCodes)

	kitchen, err := categories.Create(ctx, port.CreateCategoryParams{Name: "Cakes", Slug: "cakes", IsActive: true, Station: domaincategory.StationKitchen})
	require.NoError(t, err)
	counter, err := categories.Create(ctx, port.CreateCategoryParams{Name: "Pastries", Slug: "pastries", IsActive: true, Station: domaincategory.StationCounter})
	require.NoError(t, err)
	cake, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: kitchen.ID, Name: "Matcha cake", Slug: "matcha-cake", PriceCents: 4500, IsActive: true, LeadTime: 3 * time.Hour,
	})
	require.NoError(t, err)
	pastry, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: counter.ID, Name: "Croissant", Slug: "croissant", PriceCents: 300, IsActive: true,
	})
	require.NoError(t, err)
	combo, err := combos.Create(ctx, port.CreateComboParams{
		Name: "Tea set", Slug: "tea-set", PriceCents: 4600, IsActive: true,
		StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)
	require.NoError(t, combos.ReplaceItems(ctx, combo.ID, []port.ComboItemParams{
		{ProductID: cake.ID, Quantity: 1}, {ProductID: pastry.ID, Quantity: 1},
	}))
	f.cake, f.pastry, f.combo = cake.ID, pastry.ID, combo.ID
	return f
}

// paidPayPal stands in for PayPal with a buyer who approves every payment and
// pays it in full; a webhook call panics. taken holds the captures the shop
// never heard of, by PayPal order.
type paidPayPal struct {
	port.PaymentGateway
	mu      sync.Mutex
	amounts map[string]int64
	taken   map[string]domainpayment.Status
}

func (p *paidPayPal) CreateOrder(_ context.Context, req port.PaymentOrderRequest) (string, string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	id := "PAYPAL-" + req.OrderID.String()
	p.amounts[id] = req.AmountCents
	return id, "https://www.sandbox.paypal.com/checkoutnow?token=" + id, nil
}

// Refund returns a capture whole.
func (p *paidPayPal) Refund(_ context.Context, captureID string) (string, error) {
	return "REFUND-" + captureID, nil
}

// Lookup finds what taken holds; otherwise nothing captured, as for a buyer
// who never came back from PayPal.
func (p *paidPayPal) Lookup(_ context.Context, providerOrderID string) (*domainpayment.Capture, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	status, ok := p.taken[providerOrderID]
	if !ok {
		return nil, nil
	}
	return &domainpayment.Capture{
		ID: "CAPTURE-" + providerOrderID, Status: status,
		AmountCents: p.amounts[providerOrderID], Currency: domainpayment.Currency,
	}, nil
}

func (p *paidPayPal) Capture(_ context.Context, providerOrderID string) (domainpayment.Capture, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return domainpayment.Capture{
		ID: "CAPTURE-" + providerOrderID, Status: domainpayment.StatusCaptured,
		AmountCents: p.amounts[providerOrderID], Currency: domainpayment.Currency,
	}, nil
}

// pay pays for the customer's order the way a buyer does: approve on PayPal's
// page, come back, and the payment is captured.
func (f *checkoutFixture) pay(t *testing.T, customerID uuid.UUID, orderID string) {
	t.Helper()
	ctx := context.Background()
	_, err := f.paymentUC.Start(ctx, customerID, uuid.MustParse(orderID))
	require.NoError(t, err)
	status, err := f.paymentUC.Capture(ctx, customerID, uuid.MustParse(orderID))
	require.NoError(t, err)
	require.Equal(t, domainpayment.StatusCaptured, status)
}

// acceptingTerms is a checkout for pickupAt by a customer who accepted the
// current policies.
func acceptingTerms(pickupAt time.Time) dto.CheckoutRequest {
	return dto.CheckoutRequest{PickupAt: pickupAt, TermsVersion: domainpolicy.TermsVersion}
}

// newCustomer signs up a customer who confirmed their address, as checkout
// needs, and whose cart holds one of each line given.
func (f *checkoutFixture) newCustomer(t *testing.T, products []uuid.UUID, combos []uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	f.customer++
	customer, err := f.users.Create(ctx, port.CreateUserParams{
		Email:        fmt.Sprintf("slot-customer-%d@example.com", f.customer),
		PasswordHash: testPasswordHashFixture,
		Role:         domainuser.RoleCustomer,
	})
	require.NoError(t, err)
	require.NoError(t, f.users.MarkEmailVerified(ctx, customer.ID))
	require.NoError(t, f.fillCart(customer.ID, products, combos))
	return customer.ID
}

// fillCart puts one of each line given in the customer's cart.
func (f *checkoutFixture) fillCart(customerID uuid.UUID, products []uuid.UUID, combos []uuid.UUID) error {
	ctx := context.Background()
	for _, id := range products {
		productID := id.String()
		if _, err := f.cartUC.AddItem(ctx, customerID, dto.AddCartItemRequest{ProductID: &productID, Quantity: 1}); err != nil {
			return err
		}
	}
	for _, id := range combos {
		comboID := id.String()
		if _, err := f.cartUC.AddItem(ctx, customerID, dto.AddCartItemRequest{ComboID: &comboID, Quantity: 1}); err != nil {
			return err
		}
	}
	return nil
}

// slotNotices counts the slots.changed notices waiting in the outbox.
func (f *checkoutFixture) slotNotices(t *testing.T) int {
	t.Helper()
	count := 0
	for _, e := range claim(t, f.pool, f.outbox, 0) {
		if e.Topic == domainorder.TopicSlotsChanged {
			count++
		}
	}
	return count
}

// setSettings saves pickup rules as an admin would.
func (f *checkoutFixture) setSettings(t *testing.T, change func(*domainstore.Settings)) domainstore.Settings {
	t.Helper()
	ctx := context.Background()
	var saved domainstore.Settings
	require.NoError(t, f.pool.WithTx(ctx, func(txCtx context.Context) error {
		current, err := f.store.GetSettingsForUpdate(txCtx)
		if err != nil {
			return err
		}
		change(&current)
		if err := current.Validate(); err != nil {
			return err
		}
		saved, err = f.store.UpdateSettings(txCtx, current)
		return err
	}))
	return saved
}

// tomorrowAt is a pickup tomorrow at hour:minute bakery time: past any notice
// these tests set, inside the seeded hours and on a slot.
func tomorrowAt(hour, minute int) time.Time {
	day := domainstore.DayOf(time.Now()).AddDate(0, 0, 1)
	return time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, domainstore.Location)
}

func TestCheckoutPickup_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 8)

	t.Run("the_cart_tells_what_its_items_need", func(t *testing.T) {
		pastryOnly := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		cart, err := f.cartUC.Get(ctx, pastryOnly)
		require.NoError(t, err)
		assert.Equal(t, dto.FulfillmentResponse{HasKitchenItems: false, LeadMinutes: 0}, cart.Fulfillment)

		withCombo := f.newCustomer(t, []uuid.UUID{f.pastry}, []uuid.UUID{f.combo})
		cart, err = f.cartUC.Get(ctx, withCombo)
		require.NoError(t, err)
		assert.Equal(t, dto.FulfillmentResponse{HasKitchenItems: true, LeadMinutes: 180}, cart.Fulfillment,
			"a combo counts by the products it holds")
	})

	t.Run("a_kitchen_order_is_a_pre_order", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.cake}, nil)

		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(10, 0)))

		require.NoError(t, err)
		assert.Equal(t, string(domainorder.TypePreOrder), order.OrderType)
	})

	t.Run("a_time_between_slots_is_refused", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.cake}, nil)

		_, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(10, 10)))

		require.ErrorIs(t, err, domainorder.ErrPickupOffSlot)
		cart, err := f.cartUC.Get(ctx, customer)
		require.NoError(t, err)
		assert.Len(t, cart.Items, 1, "a refused checkout keeps the cart")
	})

	t.Run("a_full_slot_is_refused_and_a_cancelled_order_frees_it", func(t *testing.T) {
		f.setSettings(t, func(s *domainstore.Settings) { s.SlotCapacity = 1 })
		slot := tomorrowAt(11, 0)
		notices := f.slotNotices(t)
		first, err := f.orderUC.Checkout(ctx, f.newCustomer(t, []uuid.UUID{f.cake}, nil), uuid.New(), acceptingTerms(slot))
		require.NoError(t, err)
		assert.Equal(t, notices+1, f.slotNotices(t), "taking the last place tells open checkouts")

		late := f.newCustomer(t, []uuid.UUID{f.cake}, nil)
		_, err = f.orderUC.Checkout(ctx, late, uuid.New(), acceptingTerms(slot))
		require.ErrorIs(t, err, domainorder.ErrPickupSlotFull)

		_, err = f.orders.UpdateStatus(ctx, port.UpdateOrderStatusParams{
			OrderID: uuid.MustParse(first.ID), FromStatus: domainorder.StatusAwaitingPayment, ToStatus: domainorder.StatusCancelled,
		})
		require.NoError(t, err)
		_, err = f.orderUC.Checkout(ctx, late, uuid.New(), acceptingTerms(slot))
		require.NoError(t, err, "a cancelled order no longer holds its slot")
	})

	t.Run("two_checkouts_race_for_the_last_place", func(t *testing.T) {
		f.setSettings(t, func(s *domainstore.Settings) { s.SlotCapacity = 1 })
		slot := tomorrowAt(12, 0)
		racers := []uuid.UUID{f.newCustomer(t, []uuid.UUID{f.cake}, nil), f.newCustomer(t, []uuid.UUID{f.cake}, nil)}

		var wg sync.WaitGroup
		errs := make([]error, len(racers))
		start := make(chan struct{})
		for i, customer := range racers {
			wg.Go(func() {
				<-start
				_, errs[i] = f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(slot))
			})
		}
		close(start)
		wg.Wait()

		var placed, full int
		for _, err := range errs {
			switch {
			case err == nil:
				placed++
			case assert.ErrorIs(t, err, domainorder.ErrPickupSlotFull):
				full++
			}
		}
		assert.Equal(t, 1, placed, "the slot takes one order")
		assert.Equal(t, 1, full)
		booked, err := f.orders.CountByPickupTime(ctx, slot, slot.Add(time.Minute))
		require.NoError(t, err)
		require.Len(t, booked, 1)
		assert.Equal(t, 1, booked[0].Orders)
	})

	t.Run("a_slot_with_room_left_says_nothing", func(t *testing.T) {
		f.setSettings(t, func(s *domainstore.Settings) { s.SlotCapacity = 10 })
		notices := f.slotNotices(t)

		_, err := f.orderUC.Checkout(ctx, f.newCustomer(t, []uuid.UUID{f.cake}, nil), uuid.New(), acceptingTerms(tomorrowAt(13, 0)))

		require.NoError(t, err)
		assert.Equal(t, notices, f.slotNotices(t), "the slot list still shows the same")
	})

	t.Run("one_customer_books_at_most_three_orders_a_day", func(t *testing.T) {
		customer := f.newCustomer(t, nil, nil)
		cake := f.cake.String()
		for i, slot := range []time.Time{tomorrowAt(14, 0), tomorrowAt(14, 30), tomorrowAt(15, 0), tomorrowAt(15, 30)} {
			_, err := f.cartUC.AddItem(ctx, customer, dto.AddCartItemRequest{ProductID: &cake, Quantity: 1})
			require.NoError(t, err)
			_, err = f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(slot))
			if i < domainorder.MaxOrdersPerCustomerPerDay {
				require.NoError(t, err, "order %d", i+1)
				continue
			}
			require.ErrorIs(t, err, domainorder.ErrPickupDayLimit)
		}
		_, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(14, 0).AddDate(0, 0, 1)))
		require.NoError(t, err, "the next day has its own allowance")
	})

	t.Run("counter_items_collected_today_are_an_instant_order", func(t *testing.T) {
		settings := f.setSettings(t, func(s *domainstore.Settings) {
			s.OpensAt, s.ClosesAt = 0, 23*time.Hour+50*time.Minute
			s.SlotLength, s.SlotCapacity, s.InstantPrep = 10*time.Minute, 10, 0
		})
		now := time.Now().In(domainstore.Location)
		next := now.Truncate(settings.SlotLength).Add(settings.SlotLength)
		if domainstore.DayOf(next) != domainstore.DayOf(now) || next.Sub(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, domainstore.Location)) >= settings.ClosesAt {
			t.Skip("no pickup slot left today")
		}
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)

		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(next))

		require.NoError(t, err)
		assert.Equal(t, string(domainorder.TypeInstant), order.OrderType)
	})
}

// Holding a slot needs the order's transaction, and only orders that are not
// cancelled count against it.
func TestOrderRepository_PickupSlots_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 4)
	slot := tomorrowAt(9, 30)

	_, err := f.orders.HoldPickupSlot(ctx, slot, 30*time.Minute)
	require.Error(t, err, "a hold outside the order's transaction would not last")

	customer := f.newCustomer(t, nil, nil)
	for i, status := range []domainorder.Status{domainorder.StatusPending, domainorder.StatusCancelled} {
		_, err := f.orders.Create(ctx, port.CreateOrderParams{
			Code: fmt.Sprintf("CH-260101-%03d", i+1), UserID: &customer, Channel: domainorder.ChannelOnline, Status: status, Type: domainorder.TypePreOrder,
			SubtotalCents: 100, TotalCents: 100, PickupAt: &slot,
		})
		require.NoError(t, err)
	}

	var held int
	require.NoError(t, f.pool.WithTx(ctx, func(txCtx context.Context) error {
		held, err = f.orders.HoldPickupSlot(txCtx, slot, 30*time.Minute)
		return err
	}))
	assert.Equal(t, 1, held, "the cancelled order does not hold the slot")

	booked, err := f.orders.CountByPickupTime(ctx, slot.Add(-time.Hour), slot.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, booked, 1)
	assert.True(t, slot.Equal(booked[0].At))
	assert.Equal(t, 1, booked[0].Orders)
}

// An order booked before the slot grid changed still takes its place in the
// slot its pickup now falls in.
func TestCheckoutPickup_OffGridOrdersCount_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 4)
	f.setSettings(t, func(s *domainstore.Settings) { s.SlotCapacity = 1 })
	early := tomorrowAt(10, 15)
	holder := f.newCustomer(t, nil, nil)
	_, err := f.orders.Create(ctx, port.CreateOrderParams{
		Code: "CH-260101-001", UserID: &holder, Channel: domainorder.ChannelOnline, Status: domainorder.StatusConfirmed,
		Type: domainorder.TypePreOrder, SubtotalCents: 100, TotalCents: 100, PickupAt: &early,
	})
	require.NoError(t, err)

	_, err = f.orderUC.Checkout(ctx, f.newCustomer(t, []uuid.UUID{f.cake}, nil), uuid.New(), acceptingTerms(tomorrowAt(10, 0)))

	require.ErrorIs(t, err, domainorder.ErrPickupSlotFull, "10:15 sits in the 10:00 slot of the current grid")
}
