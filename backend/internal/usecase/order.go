package usecase

import (
	"context"
	"errors"
	"time"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
)

const (
	OrderListDefaultPageSize      int32 = 20
	OrderListMaxPageSize          int32 = 100
	OrderListDefaultPageSizeQuery       = "20"
)

type OrderUsecase struct {
	users       port.UserRepository
	orders      port.OrderRepository
	carts       port.CartRepository
	discount    port.DiscountCodeRepository
	cartUC      *CartUsecase
	tx          port.TxManager
	events      port.EventOutbox
	store       port.StoreSettingsRepository
	tickets     port.TicketRepository
	payments    port.PaymentRepository
	payment     *PaymentUsecase
	pickupCodes domainorder.PickupCodes
	transitions orderTransitions
}

func NewOrderUsecase(
	users port.UserRepository,
	orders port.OrderRepository,
	carts port.CartRepository,
	discount port.DiscountCodeRepository,
	cartUC *CartUsecase,
	tx port.TxManager,
	events port.EventOutbox,
	store port.StoreSettingsRepository,
	tickets port.TicketRepository,
	payments port.PaymentRepository,
	payment *PaymentUsecase,
	pickupCodes domainorder.PickupCodes,
) *OrderUsecase {
	return &OrderUsecase{
		users: users, orders: orders, carts: carts, discount: discount, cartUC: cartUC,
		tx: tx, events: events, store: store, tickets: tickets, payments: payments, payment: payment,
		pickupCodes: pickupCodes,
		transitions: orderTransitions{
			tx: tx, orders: orders, tickets: tickets, discounts: discount, payments: payments, events: events,
		},
	}
}

// Checkout places the customer's cart as an order awaiting payment: it holds
// its pickup slot and discount until paid. checkoutKey is the request's
// Idempotency-Key: the same key again returns the order it placed, so a retry
// after a lost answer places nothing twice.
func (u *OrderUsecase) Checkout(ctx context.Context, userID, checkoutKey uuid.UUID, req dto.CheckoutRequest) (*dto.OrderResponse, error) {
	if err := domainpolicy.RequireAccepted(req.TermsVersion); err != nil {
		return nil, err
	}
	if placed, err := u.orders.GetByCheckoutKey(ctx, userID, checkoutKey); err == nil {
		return u.orderResponse(ctx, userID, placed.ID)
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	pickupAt := req.PickupAt
	now := time.Now()
	settings, closed, err := readPickupRules(ctx, u.store, now)
	if err != nil {
		return nil, err
	}
	policy := pickupPolicy(settings, closed)

	var created *domainorder.Order
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		cart, lines, discountCode, totals, err := u.cartUC.pricedCartForCheckout(txCtx, userID)
		if err != nil {
			return err
		}
		// Taken after the cart, the order an erasure takes them in, so an erasure
		// or a disable waits for this order. The other way round, an erasure
		// holds the cart and this checkout finds it emptied; a cart filled after
		// the erasure, or a disable, leaves the account closed here.
		customer, err := u.users.GetByIDForShare(txCtx, userID)
		if errors.Is(err, apperrors.ErrNotFound) {
			return ErrMeNotFound
		}
		if err != nil {
			return err
		}
		// Every update about the order goes to this address.
		if !customer.EmailVerified {
			return domainuser.ErrEmailNotVerified
		}
		items, err := u.cartUC.fulfillmentOf(txCtx, lines)
		if err != nil {
			return err
		}
		booking, err := bookPickup(txCtx, u.orders, &userID, items, pickupAt, now, policy, nil)
		if err != nil {
			return err
		}

		var discountCodeID *uuid.UUID
		var discountSnapshot *string
		if discountCode != nil {
			if err := u.holdDiscount(txCtx, userID, discountCode.ID); err != nil {
				return err
			}
			discountCodeID = &discountCode.ID
			snapshot := discountCode.Code
			discountSnapshot = &snapshot
		}

		// Taken last before the insert: the day's number stays locked until commit.
		day, number, err := u.orders.NextDayNumber(txCtx)
		if err != nil {
			return err
		}
		order, err := u.orders.Create(txCtx, port.CreateOrderParams{
			UserID:               &userID,
			Channel:              domainorder.ChannelOnline,
			Code:                 domainorder.Code(day, number),
			Status:               domainorder.StatusAwaitingPayment,
			Type:                 booking.orderType,
			SubtotalCents:        totals.SubtotalCents,
			DiscountCents:        totals.DiscountCents,
			TotalCents:           totals.TotalCents,
			DiscountCodeID:       discountCodeID,
			DiscountCodeSnapshot: discountSnapshot,
			PickupAt:             &pickupAt,
			TermsVersion:         &req.TermsVersion,
			CheckoutKey:          &checkoutKey,
			PaymentHold:          settings.PaymentHold,
		})
		if err != nil {
			return err
		}

		if err := u.orders.CreateItems(txCtx, orderItemsOf(order.ID, lines)); err != nil {
			return err
		}
		if err := createOrderTickets(txCtx, u.tickets, order.ID); err != nil {
			return err
		}
		if err := u.orders.AddStatusEvent(txCtx, port.AddOrderStatusEventParams{
			OrderID: order.ID,
			To:      order.Status,
			Actor:   &port.OrderActor{ID: userID, Role: domainuser.RoleCustomer},
		}); err != nil {
			return err
		}
		if err := u.carts.DeleteAllItems(txCtx, cart.ID); err != nil {
			return err
		}
		if err := u.carts.ClearDiscountCode(txCtx, cart.ID); err != nil {
			return err
		}
		if err := u.events.Add(txCtx, domainorder.CreatedEvent(*order)); err != nil {
			return err
		}
		// Nothing to pay, as when a discount covers the whole order: it is paid
		// at once and goes on the way a paid order does.
		if order.TotalCents == 0 {
			customer := &port.OrderActor{ID: userID, Role: domainuser.RoleCustomer}
			if order, err = u.transitions.applyInTx(txCtx, customer, port.UpdateOrderStatusParams{
				OrderID:    order.ID,
				FromStatus: domainorder.StatusAwaitingPayment,
				ToStatus:   domainorder.PaidStatus(hasCustomLine(lines)),
			}, ""); err != nil {
				return err
			}
		}
		if booking.fillsSlot {
			if err := u.events.Add(txCtx, domainorder.SlotsChangedEvent(domainstore.DayOf(pickupAt))); err != nil {
				return err
			}
		}
		created = order
		return nil
	})
	if errors.Is(err, domaincart.ErrEmpty) || errors.Is(err, apperrors.ErrConflict) {
		// A request with the same key placed the order while this one waited
		// on the cart.
		if placed, lookupErr := u.orders.GetByCheckoutKey(ctx, userID, checkoutKey); lookupErr == nil {
			return u.orderResponse(ctx, userID, placed.ID)
		}
	}
	if err != nil {
		return nil, err
	}
	return u.orderResponse(ctx, userID, created.ID)
}

// holdDiscount takes one of the code's uses for the order being placed. The
// code's row stays locked until the order commits, so orders placed at the
// same time count each other: neither the code's cap nor its cap per customer
// is passed.
func (u *OrderUsecase) holdDiscount(txCtx context.Context, userID, codeID uuid.UUID) error {
	code, err := u.discount.IncrementUsedCount(txCtx, codeID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return domaindiscount.ErrExhausted
	}
	if err != nil || code.MaxUsesPerCustomer == nil {
		return err
	}
	used, err := u.orders.CountCustomerDiscountUses(txCtx, userID, codeID)
	if err != nil {
		return err
	}
	if used >= int(*code.MaxUsesPerCustomer) {
		return domaindiscount.ErrUsedUpByCustomer
	}
	return nil
}

// orderItemsOf turns priced lines into a new order's receipt lines.
func orderItemsOf(orderID uuid.UUID, lines []pricedCartLine) []port.CreateOrderItemParams {
	items := make([]port.CreateOrderItemParams, 0, len(lines))
	for _, line := range lines {
		items = append(items, port.CreateOrderItemParams{
			OrderID:        orderID,
			LineType:       line.Item.LineType,
			Customization:  line.Customization,
			Name:           line.Name,
			Slug:           line.Slug,
			Quantity:       line.Item.Quantity,
			UnitPriceCents: line.UnitPriceCents,
			LineTotalCents: line.LineTotalCents,
			ProductID:      line.Item.ProductID,
			ComboID:        line.Item.ComboID,
		})
	}
	return items
}

// createOrderTickets splits a new order into one ticket per station, combos by
// the products they hold.
func createOrderTickets(txCtx context.Context, tickets port.TicketRepository, orderID uuid.UUID) error {
	lines, err := tickets.ListLines(txCtx, orderID)
	if err != nil {
		return err
	}
	_, err = tickets.CreateForOrder(txCtx, orderID, domainorder.Decompose(lines))
	return err
}

// pickupBooking is what booking a pickup decided: the order's type, and
// whether the order takes the slot's last place.
type pickupBooking struct {
	orderType domainorder.Type
	fillsSlot bool
}

// bookPickup decides the order's type from what its items ask of the bakery,
// checks the pickup time against the rules for that type, and holds its slot
// until the transaction ends. It refuses a slot that is already full and a
// customer who already holds as many orders for that day as one may (a guest,
// nil customer, has no account to count); current is the pickup an order being
// moved holds now, nil for a new order. Both holds last until the transaction
// ends, so neither count can race.
func bookPickup(
	txCtx context.Context,
	orders port.OrderRepository,
	customer *uuid.UUID,
	items domainorder.Fulfillment,
	pickupAt, now time.Time,
	policy domainorder.PickupPolicy,
	current *time.Time,
) (pickupBooking, error) {
	day := domainstore.DayOf(pickupAt)
	// An order moving within its own day holds that day already: it counts
	// toward it, and its items are its own though one has sold out since.
	sameDay := current != nil && domainstore.DayOf(*current).Equal(day)
	if sameDay {
		items = items.HeldOn(day)
	}
	orderType, err := policy.Validate(pickupAt, now, items)
	if err != nil {
		return pickupBooking{}, err
	}
	if customer != nil && !sameDay {
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, domainstore.Location)
		mine, err := orders.HoldCustomerDay(txCtx, *customer, dayStart, dayStart.AddDate(0, 0, 1))
		if err != nil {
			return pickupBooking{}, err
		}
		if mine >= domainorder.MaxOrdersPerCustomerPerDay {
			return pickupBooking{}, domainorder.ErrPickupDayLimit
		}
	}
	held, err := orders.HoldPickupSlot(txCtx, pickupAt, policy.Settings.SlotLength)
	if err != nil {
		return pickupBooking{}, err
	}
	if held >= policy.Settings.SlotCapacity {
		return pickupBooking{}, domainorder.ErrPickupSlotFull
	}
	return pickupBooking{orderType: orderType, fillsSlot: held+1 == policy.Settings.SlotCapacity}, nil
}

// List returns a page of the customer's orders, newest first, narrowed by query.
func (u *OrderUsecase) List(
	ctx context.Context,
	userID uuid.UUID,
	page, pageSize int32,
	query dto.OrderHistoryQuery,
) ([]dto.OrderSummaryResponse, int64, int32, int32, error) {
	page, pageSize = normalizeOrderListPage(page, pageSize)
	filter, err := orderHistoryFilter(query)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	orders, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]domainorder.Order, error) {
			return u.orders.ListByUser(ctx, port.ListOrdersParams{
				UserID: userID,
				Filter: filter,
				Limit:  pageSize,
				Offset: utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.orders.ListCountByUser(ctx, userID, filter)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	orderIDs := make([]uuid.UUID, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}
	itemCounts, err := u.orders.SumItemQuantitiesByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.OrderSummaryResponse, 0, len(orders))
	for _, order := range orders {
		out = append(out, dto.OrderSummaryResponse{
			ID:         order.ID.String(),
			Code:       order.Code,
			Status:     string(order.Status),
			TotalCents: order.TotalCents,
			ItemCount:  itemCounts[order.ID],
			PickupAt:   order.PickupAt,
			CreatedAt:  order.CreatedAt,
		})
	}
	return out, total, page, pageSize, nil
}

func (u *OrderUsecase) Get(ctx context.Context, userID, orderID uuid.UUID) (*dto.OrderResponse, error) {
	return u.orderResponse(ctx, userID, orderID)
}

func (u *OrderUsecase) orderResponse(ctx context.Context, userID, orderID uuid.UUID) (*dto.OrderResponse, error) {
	order, err := u.orders.GetByIDForUser(ctx, userID, orderID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainorder.ErrNotFound
		}
		return nil, err
	}
	parts, err := readOrderDetail(ctx, u.orders, u.tickets, u.payments, order.ID)
	if err != nil {
		return nil, err
	}
	resp := &dto.OrderResponse{
		ID:                   order.ID.String(),
		Code:                 order.Code,
		Status:               string(order.Status),
		OrderType:            string(order.Type),
		SubtotalCents:        order.SubtotalCents,
		DiscountCents:        order.DiscountCents,
		TotalCents:           order.TotalCents,
		DiscountCodeSnapshot: order.DiscountCodeSnapshot,
		PickupAt:             order.PickupAt,
		Items:                mapOrderItemsToDTO(parts.items),
		Timeline:             mapOrderTimelineToDTO(parts.timeline),
		Tickets:              mapTicketSummariesToDTO(parts.tickets),
		Payment:              mapOrderPaymentToDTO(parts.payment),
		PaymentDueAt:         order.PaymentDueAt,
		CreatedAt:            order.CreatedAt,
		UpdatedAt:            order.UpdatedAt,
	}
	// The pickup picker needs what the items ask of the bakery only while the
	// customer may still move the pickup.
	if order.Status.BeforeProduction() {
		items, err := u.cartUC.orderFulfillment(ctx, parts.items)
		if err != nil {
			return nil, err
		}
		if order.PickupAt != nil {
			items = items.HeldOn(domainstore.DayOf(*order.PickupAt))
		}
		fulfillment := mapFulfillmentToDTO(items)
		resp.Fulfillment = &fulfillment
	}
	if order.HasPickupCode() {
		code := u.pickupCodes.Of(order.ID)
		resp.PickupCode = &code
	}
	return resp, nil
}

func normalizeOrderListPage(page, pageSize int32) (int32, int32) {
	return utils.NormalizePageParams(page, pageSize, OrderListDefaultPageSize, OrderListMaxPageSize)
}
