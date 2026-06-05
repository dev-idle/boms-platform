package postgres

import (
	"context"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaincart "github.com/boms/backend/internal/domain/cart"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/google/uuid"
)

type CartRepository struct {
	queries *sqlcgen.Queries
}

func NewCartRepository(pool *Pool) *CartRepository {
	return &CartRepository{queries: pool.Queries()}
}

func (r *CartRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *CartRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domaincart.Cart, error) {
	row, err := r.q(ctx).GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, mapRepoError(err, "get cart by user")
	}
	return mapCart(row), nil
}

func (r *CartRepository) CreateForUser(ctx context.Context, userID uuid.UUID) (*domaincart.Cart, error) {
	row, err := r.q(ctx).CreateCart(ctx, userID)
	if err != nil {
		return nil, mapRepoError(err, "create cart")
	}
	return mapCart(row), nil
}

func (r *CartRepository) SetDiscountCodeID(ctx context.Context, cartID uuid.UUID, discountCodeID *uuid.UUID) error {
	err := r.q(ctx).SetCartDiscountCodeID(ctx, sqlcgen.SetCartDiscountCodeIDParams{
		ID:             cartID,
		DiscountCodeID: optionalUUID(discountCodeID),
	})
	if err != nil {
		return mapRepoError(err, "set cart discount code")
	}
	return nil
}

func (r *CartRepository) ClearDiscountCode(ctx context.Context, cartID uuid.UUID) error {
	return r.SetDiscountCodeID(ctx, cartID, nil)
}

func (r *CartRepository) ListItemsByCartID(ctx context.Context, cartID uuid.UUID) ([]domaincart.Item, error) {
	rows, err := r.q(ctx).ListCartItemsByCartID(ctx, cartID)
	if err != nil {
		return nil, mapRepoError(err, "list cart items")
	}
	out := make([]domaincart.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCartItem(row))
	}
	return out, nil
}

func (r *CartRepository) GetItemByID(ctx context.Context, cartID, itemID uuid.UUID) (*domaincart.Item, error) {
	row, err := r.q(ctx).GetCartItemByID(ctx, sqlcgen.GetCartItemByIDParams{
		CartID: cartID,
		ID:     itemID,
	})
	if err != nil {
		return nil, mapRepoError(err, "get cart item")
	}
	item := mapCartItem(row)
	return &item, nil
}

func (r *CartRepository) GetItemByProduct(ctx context.Context, cartID, productID uuid.UUID) (*domaincart.Item, error) {
	row, err := r.q(ctx).GetCartItemByProduct(ctx, sqlcgen.GetCartItemByProductParams{
		CartID:    cartID,
		ProductID: uuid.NullUUID{UUID: productID, Valid: true},
	})
	if err != nil {
		return nil, mapRepoError(err, "get cart item by product")
	}
	item := mapCartItem(row)
	return &item, nil
}

func (r *CartRepository) GetItemByCombo(ctx context.Context, cartID, comboID uuid.UUID) (*domaincart.Item, error) {
	row, err := r.q(ctx).GetCartItemByCombo(ctx, sqlcgen.GetCartItemByComboParams{
		CartID:  cartID,
		ComboID: uuid.NullUUID{UUID: comboID, Valid: true},
	})
	if err != nil {
		return nil, mapRepoError(err, "get cart item by combo")
	}
	item := mapCartItem(row)
	return &item, nil
}

func (r *CartRepository) CreateItem(ctx context.Context, params port.CreateCartItemParams) (*domaincart.Item, error) {
	lineType, err := mapCartLineTypeToSQL(params.LineType)
	if err != nil {
		return nil, err
	}
	row, err := r.q(ctx).CreateCartItem(ctx, sqlcgen.CreateCartItemParams{
		CartID:    params.CartID,
		LineType:  lineType,
		ProductID: optionalUUID(params.ProductID),
		ComboID:   optionalUUID(params.ComboID),
		Quantity:  params.Quantity,
	})
	if err != nil {
		return nil, mapRepoError(err, "create cart item")
	}
	item := mapCartItem(row)
	return &item, nil
}

func (r *CartRepository) UpdateItemQuantity(
	ctx context.Context,
	params port.UpdateCartItemQuantityParams,
) (*domaincart.Item, error) {
	row, err := r.q(ctx).UpdateCartItemQuantity(ctx, sqlcgen.UpdateCartItemQuantityParams{
		CartID:   params.CartID,
		ID:       params.ID,
		Quantity: params.Quantity,
	})
	if err != nil {
		return nil, mapRepoError(err, "update cart item quantity")
	}
	item := mapCartItem(row)
	return &item, nil
}

func (r *CartRepository) DeleteItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	rows, err := r.q(ctx).DeleteCartItem(ctx, sqlcgen.DeleteCartItemParams{
		CartID: cartID,
		ID:     itemID,
	})
	if err != nil {
		return mapRepoError(err, "delete cart item")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *CartRepository) DeleteAllItems(ctx context.Context, cartID uuid.UUID) error {
	if err := r.q(ctx).DeleteAllCartItems(ctx, cartID); err != nil {
		return mapRepoError(err, "delete all cart items")
	}
	return nil
}

func (r *CartRepository) CountItems(ctx context.Context, cartID uuid.UUID) (int64, error) {
	count, err := r.q(ctx).CountCartItems(ctx, cartID)
	if err != nil {
		return 0, mapRepoError(err, "count cart items")
	}
	return count, nil
}

func mapCart(row sqlcgen.Cart) *domaincart.Cart {
	c := &domaincart.Cart{
		ID:        row.ID,
		UserID:    row.UserID,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.DiscountCodeID.Valid {
		id := row.DiscountCodeID.UUID
		c.DiscountCodeID = &id
	}
	return c
}

func mapCartItem(row sqlcgen.CartItem) domaincart.Item {
	item := domaincart.Item{
		ID:        row.ID,
		CartID:    row.CartID,
		LineType:  mapCartLineTypeFromSQL(row.LineType),
		Quantity:  row.Quantity,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.ProductID.Valid {
		id := row.ProductID.UUID
		item.ProductID = &id
	}
	if row.ComboID.Valid {
		id := row.ComboID.UUID
		item.ComboID = &id
	}
	return item
}

func mapCartLineTypeToSQL(t domaincart.LineType) (sqlcgen.CartLineType, error) {
	switch t {
	case domaincart.LineTypeProduct:
		return sqlcgen.CartLineTypeProduct, nil
	case domaincart.LineTypeCombo:
		return sqlcgen.CartLineTypeCombo, nil
	default:
		return "", apperrors.Errorf("unsupported cart line type: %s", t)
	}
}

func mapCartLineTypeFromSQL(t sqlcgen.CartLineType) domaincart.LineType {
	switch t {
	case sqlcgen.CartLineTypeProduct:
		return domaincart.LineTypeProduct
	case sqlcgen.CartLineTypeCombo:
		return domaincart.LineTypeCombo
	default:
		return domaincart.LineType(t)
	}
}
