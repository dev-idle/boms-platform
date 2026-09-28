package usecase

import (
	"context"
	"errors"
	"time"

	domainorder "github.com/boms/backend/internal/domain/order"
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
	orders   port.OrderRepository
	carts    port.CartRepository
	discount port.DiscountCodeRepository
	cartUC   *CartUsecase
	tx       port.TxManager
	events   port.EventOutbox
	store    port.StoreSettingsRepository
}

func NewOrderUsecase(
	orders port.OrderRepository,
	carts port.CartRepository,
	discount port.DiscountCodeRepository,
	cartUC *CartUsecase,
	tx port.TxManager,
	events port.EventOutbox,
	store port.StoreSettingsRepository,
) *OrderUsecase {
	return &OrderUsecase{orders: orders, carts: carts, discount: discount, cartUC: cartUC, tx: tx, events: events, store: store}
}

func (u *OrderUsecase) Checkout(ctx context.Context, userID uuid.UUID, pickupAt time.Time) (*dto.OrderResponse, error) {
	now := time.Now()
	settings, closed, err := readPickupRules(ctx, u.store, now)
	if err != nil {
		return nil, err
	}
	if err := pickupPolicy(settings, closed).Validate(pickupAt, now); err != nil {
		return nil, err
	}

	var created *domainorder.Order
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		cart, lines, discountCode, totals, err := u.cartUC.pricedCartForCheckout(txCtx, userID)
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
			SubtotalCents:        totals.SubtotalCents,
			DiscountCents:        totals.DiscountCents,
			TotalCents:           totals.TotalCents,
			DiscountCodeID:       discountCodeID,
			DiscountCodeSnapshot: discountSnapshot,
			PickupAt:             &pickupAt,
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
		created = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return u.orderResponse(ctx, userID, created.ID)
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
	items, timeline, err := orderLinesAndTimeline(ctx, u.orders, order.ID)
	if err != nil {
		return nil, err
	}
	resp := &dto.OrderResponse{
		ID:                   order.ID.String(),
		Code:                 order.Code,
		Status:               string(order.Status),
		SubtotalCents:        order.SubtotalCents,
		DiscountCents:        order.DiscountCents,
		TotalCents:           order.TotalCents,
		DiscountCodeSnapshot: order.DiscountCodeSnapshot,
		PickupAt:             order.PickupAt,
		Items:                mapOrderItemsToDTO(items),
		Timeline:             mapOrderTimelineToDTO(timeline),
		CreatedAt:            order.CreatedAt,
		UpdatedAt:            order.UpdatedAt,
	}
	return resp, nil
}

func normalizeOrderListPage(page, pageSize int32) (int32, int32) {
	return utils.NormalizePageParams(page, pageSize, OrderListDefaultPageSize, OrderListMaxPageSize)
}
