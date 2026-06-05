package usecase

import (
	"context"
	"errors"

	domainorder "github.com/boms/backend/internal/domain/order"
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
}

func NewOrderUsecase(
	orders port.OrderRepository,
	carts port.CartRepository,
	discount port.DiscountCodeRepository,
	cartUC *CartUsecase,
	tx port.TxManager,
) *OrderUsecase {
	return &OrderUsecase{orders: orders, carts: carts, discount: discount, cartUC: cartUC, tx: tx}
}

func (u *OrderUsecase) Checkout(ctx context.Context, userID uuid.UUID) (*dto.OrderResponse, error) {
	var created *domainorder.Order
	err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
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

		order, err := u.orders.Create(txCtx, port.CreateOrderParams{
			UserID:               userID,
			Status:               domainorder.StatusPending,
			SubtotalCents:        totals.SubtotalCents,
			DiscountCents:        totals.DiscountCents,
			TotalCents:           totals.TotalCents,
			DiscountCodeID:       discountCodeID,
			DiscountCodeSnapshot: discountSnapshot,
		})
		if err != nil {
			return err
		}

		orderItems := make([]port.CreateOrderItemParams, 0, len(lines))
		for _, line := range lines {
			params := port.CreateOrderItemParams{
				OrderID:        order.ID,
				LineType:       line.Item.LineType,
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
		created = order
		return nil
	})
	if err != nil {
		return nil, err
	}
	return u.orderResponse(ctx, userID, created.ID)
}

func (u *OrderUsecase) List(
	ctx context.Context,
	userID uuid.UUID,
	page, pageSize int32,
) ([]dto.OrderSummaryResponse, int64, int32, int32, error) {
	page, pageSize = normalizeOrderListPage(page, pageSize)
	orders, err := u.orders.ListByUser(ctx, port.ListOrdersParams{
		UserID: userID,
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.orders.ListCountByUser(ctx, userID)
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
			Status:     string(order.Status),
			TotalCents: order.TotalCents,
			ItemCount:  itemCounts[order.ID],
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
	items, err := u.orders.ListItemsByOrderID(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	resp := &dto.OrderResponse{
		ID:                   order.ID.String(),
		Status:               string(order.Status),
		SubtotalCents:        order.SubtotalCents,
		DiscountCents:        order.DiscountCents,
		TotalCents:           order.TotalCents,
		DiscountCodeSnapshot: order.DiscountCodeSnapshot,
		Items:                make([]dto.OrderItemResponse, 0, len(items)),
		CreatedAt:            order.CreatedAt,
		UpdatedAt:            order.UpdatedAt,
	}
	for _, item := range items {
		row := dto.OrderItemResponse{
			ID:             item.ID.String(),
			LineType:       string(item.LineType),
			Name:           item.Name,
			Slug:           item.Slug,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: item.LineTotalCents,
		}
		if item.ProductID != nil {
			id := item.ProductID.String()
			row.ProductID = &id
		}
		if item.ComboID != nil {
			id := item.ComboID.String()
			row.ComboID = &id
		}
		resp.Items = append(resp.Items, row)
	}
	return resp, nil
}

func normalizeOrderListPage(page, pageSize int32) (int32, int32) {
	return utils.NormalizePageParams(page, pageSize, OrderListDefaultPageSize, OrderListMaxPageSize)
}
