package port

import (
	"context"

	domaincart "github.com/boms/backend/internal/domain/cart"
	"github.com/google/uuid"
)

type CreateCartItemParams struct {
	CartID    uuid.UUID
	LineType  domaincart.LineType
	ProductID *uuid.UUID
	ComboID   *uuid.UUID
	Quantity  int32
}

type UpdateCartItemQuantityParams struct {
	ID       uuid.UUID
	CartID   uuid.UUID
	Quantity int32
}

type CartRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domaincart.Cart, error)
	CreateForUser(ctx context.Context, userID uuid.UUID) (*domaincart.Cart, error)
	SetDiscountCodeID(ctx context.Context, cartID uuid.UUID, discountCodeID *uuid.UUID) error
	ClearDiscountCode(ctx context.Context, cartID uuid.UUID) error
	ListItemsByCartID(ctx context.Context, cartID uuid.UUID) ([]domaincart.Item, error)
	GetItemByID(ctx context.Context, cartID, itemID uuid.UUID) (*domaincart.Item, error)
	GetItemByProduct(ctx context.Context, cartID, productID uuid.UUID) (*domaincart.Item, error)
	GetItemByCombo(ctx context.Context, cartID, comboID uuid.UUID) (*domaincart.Item, error)
	CreateItem(ctx context.Context, params CreateCartItemParams) (*domaincart.Item, error)
	UpdateItemQuantity(ctx context.Context, params UpdateCartItemQuantityParams) (*domaincart.Item, error)
	DeleteItem(ctx context.Context, cartID, itemID uuid.UUID) error
	DeleteAllItems(ctx context.Context, cartID uuid.UUID) error
	CountItems(ctx context.Context, cartID uuid.UUID) (int64, error)
}
