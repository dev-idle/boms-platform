package usecase

import (
	"context"
	"errors"
	"time"

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
	users    port.UserRepository
	orders   port.OrderRepository
	carts    port.CartRepository
	discount port.DiscountCodeRepository
	cartUC   *CartUsecase
	tx       port.TxManager
	events   port.EventOutbox
	store    port.StoreSettingsRepository
	tickets  port.TicketRepository
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
) *OrderUsecase {
	return &OrderUsecase{
		users: users, orders: orders, carts: carts, discount: discount, cartUC: cartUC,
		tx: tx, events: events, store: store, tickets: tickets,
	}
}

func (u *OrderUsecase) Checkout(ctx context.Context, userID uuid.UUID, req dto.CheckoutRequest) (*dto.OrderResponse, error) {
	if err := domainpolicy.RequireAccepted(req.TermsVersion); err != nil {
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
		booking, err := u.bookPickup(txCtx, userID, lines, pickupAt, now, policy)
		if err != nil {
			return err
		}

		var discountCodeID *uuid.UUID
		var discountSnapshot *string
		if discountCode != nil {
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
			UserID:               userID,
			Code:                 domainorder.Code(day, number),
			Status:               domainorder.StatusPending,
			Type:                 booking.orderType,
			SubtotalCents:        totals.SubtotalCents,
			DiscountCents:        totals.DiscountCents,
			TotalCents:           totals.TotalCents,
			DiscountCodeID:       discountCodeID,
			DiscountCodeSnapshot: discountSnapshot,
			PickupAt:             &pickupAt,
			TermsVersion:         &req.TermsVersion,
		})
		if err != nil {
			return err
		}

		orderItems := make([]port.CreateOrderItemParams, 0, len(lines))
		for _, line := range lines {
			params := port.CreateOrderItemParams{
				OrderID:        order.ID,
				LineType:       line.Item.LineType,
				Configuration:  line.Item.Configuration,
				Name:           line.Name,
				Slug:           line.Slug,
				Quantity:       line.Item.Quantity,
				UnitPriceCents: line.UnitPriceCents,
				LineTotalCents: line.LineTotalCents,
				ProductID:      line.Item.ProductID,
				ComboID:        line.Item.ComboID,
			}
			orderItems = append(orderItems, params)
		}
		if err := u.orders.CreateItems(txCtx, orderItems); err != nil {
			return err
		}
		if err := u.createTickets(txCtx, order.ID); err != nil {
			return err
		}
		if err := u.orders.AddStatusEvent(txCtx, port.AddOrderStatusEventParams{
			OrderID:   order.ID,
			To:        order.Status,
			ActorID:   userID,
			ActorRole: domainuser.RoleCustomer,
		}); err != nil {
			return err
		}
		if discountCode != nil {
			if _, err := u.discount.IncrementUsedCount(txCtx, discountCode.ID); err != nil {
				return err
			}
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
		if booking.fillsSlot {
			if err := u.events.Add(txCtx, domainorder.SlotsChangedEvent(domainstore.DayOf(pickupAt))); err != nil {
				return err
			}
		}
		created = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return u.orderResponse(ctx, userID, created.ID)
}

// createTickets splits a new order into one ticket per station, combos by the
// products they hold.
func (u *OrderUsecase) createTickets(txCtx context.Context, orderID uuid.UUID) error {
	lines, err := u.tickets.ListLines(txCtx, orderID)
	if err != nil {
		return err
	}
	_, err = u.tickets.CreateForOrder(txCtx, orderID, domainorder.Decompose(lines))
	return err
}

// pickupBooking is what booking a pickup decided: the order's type, and
// whether the order takes the slot's last place.
type pickupBooking struct {
	orderType domainorder.Type
	fillsSlot bool
}

// bookPickup decides the order's type from the priced lines of the locked
// cart, checks the pickup time against the rules for that type, and holds its
// slot until the transaction ends. It refuses a slot that is already full and
// a customer who already holds as many orders for that day as one may. The
// cart lock serializes one customer's checkouts, so the day count cannot race.
func (u *OrderUsecase) bookPickup(
	txCtx context.Context,
	userID uuid.UUID,
	lines []pricedCartLine,
	pickupAt, now time.Time,
	policy domainorder.PickupPolicy,
) (pickupBooking, error) {
	items, err := u.cartUC.fulfillmentOf(txCtx, lines)
	if err != nil {
		return pickupBooking{}, err
	}
	orderType, err := policy.Validate(pickupAt, now, items)
	if err != nil {
		return pickupBooking{}, err
	}
	day := domainstore.DayOf(pickupAt)
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, domainstore.Location)
	mine, err := u.orders.CountCustomerOrdersBetween(txCtx, userID, dayStart, dayStart.AddDate(0, 0, 1))
	if err != nil {
		return pickupBooking{}, err
	}
	if mine >= domainorder.MaxOrdersPerCustomerPerDay {
		return pickupBooking{}, domainorder.ErrPickupDayLimit
	}
	held, err := u.orders.HoldPickupSlot(txCtx, pickupAt, policy.Settings.SlotLength)
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
	parts, err := readOrderDetail(ctx, u.orders, u.tickets, order.ID)
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
		CreatedAt:            order.CreatedAt,
		UpdatedAt:            order.UpdatedAt,
	}
	return resp, nil
}

func normalizeOrderListPage(page, pageSize int32) (int32, int32) {
	return utils.NormalizePageParams(page, pageSize, OrderListDefaultPageSize, OrderListMaxPageSize)
}
