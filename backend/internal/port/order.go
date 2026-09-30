package port

import (
	"context"
	"encoding/json"
	"time"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/google/uuid"
)

type CreateOrderParams struct {
	UserID               uuid.UUID
	Code                 string
	Status               domainorder.Status
	Type                 domainorder.Type
	SubtotalCents        int64
	DiscountCents        int64
	TotalCents           int64
	DiscountCodeID       *uuid.UUID
	DiscountCodeSnapshot *string
	PickupAt             *time.Time
	// TermsVersion is the policy version the customer accepted; the order
	// records it with the instant it was placed.
	TermsVersion *string
	// CheckoutKey is the Idempotency-Key of the checkout placing the order.
	CheckoutKey *uuid.UUID
	// PaymentHold is how long the order waits for payment before it expires.
	PaymentHold time.Duration
}

type CreateOrderItemParams struct {
	OrderID        uuid.UUID
	LineType       domaincart.LineType
	ProductID      *uuid.UUID
	ComboID        *uuid.UUID
	Configuration  json.RawMessage
	Name           string
	Slug           string
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
}

// OrderHistoryFilter narrows a customer's order history. Nil fields do not
// filter; PlacedBefore is exclusive.
type OrderHistoryFilter struct {
	Status       *domainorder.Status
	PlacedFrom   *time.Time
	PlacedBefore *time.Time
}

type ListOrdersParams struct {
	UserID uuid.UUID
	Filter OrderHistoryFilter
	Limit  int32
	Offset int32
}

type StaffListOrdersParams struct {
	Status *domainorder.Status
	Limit  int32
	Offset int32
}

type StaffOrderListRow struct {
	Order                 domainorder.Order
	CustomerEmail         string
	CustomerEmailVerified bool
	CustomerDisplayName   *string
	CustomerPhone         *string
}

// PickupCount is how many orders not cancelled are due at one pickup time.
type PickupCount struct {
	At     time.Time
	Orders int
}

// OrderActor is the signed-in user who moved an order, in the role they did it in.
type OrderActor struct {
	ID   uuid.UUID
	Role domainuser.Role
}

// AddOrderStatusEventParams records an order entering To. From is nil when the
// order is being placed; Actor is nil when the system made the move, as when
// an unpaid order expires.
type AddOrderStatusEventParams struct {
	OrderID uuid.UUID
	From    *domainorder.Status
	To      domainorder.Status
	Actor   *OrderActor
}

type UpdateOrderStatusParams struct {
	OrderID    uuid.UUID
	FromStatus domainorder.Status
	ToStatus   domainorder.Status
}

type OrderRepository interface {
	Create(ctx context.Context, params CreateOrderParams) (*domainorder.Order, error)
	GetByIDForUser(ctx context.Context, userID, orderID uuid.UUID) (*domainorder.Order, error)
	// GetByCheckoutKey returns the customer's order placed by the checkout
	// with that Idempotency-Key.
	GetByCheckoutKey(ctx context.Context, userID, checkoutKey uuid.UUID) (*domainorder.Order, error)
	// CountCustomerDiscountUses counts the customer's orders not cancelled or
	// expired that use the discount code.
	CountCustomerDiscountUses(ctx context.Context, userID, discountCodeID uuid.UUID) (int, error)
	// ListDueUnpaid returns up to limit orders awaiting payment more than
	// grace past their due time, the longest overdue first.
	ListDueUnpaid(ctx context.Context, grace time.Duration, limit int32) ([]uuid.UUID, error)
	// Expire expires an order still awaiting payment more than grace past its
	// due time. Any other is apperrors.ErrNotFound: it was paid meanwhile.
	Expire(ctx context.Context, orderID uuid.UUID, grace time.Duration) (*domainorder.Order, error)
	// LockForUpdate row-locks the order until the transaction ends; call it only
	// inside one. Every ticket move takes it first; the order's own status moves
	// take the same row lock through their guarded UPDATE.
	LockForUpdate(ctx context.Context, orderID uuid.UUID) (*domainorder.Order, error)
	// StaffGetByID returns an order with its customer's contact details. An
	// order whose customer's account is closed (disabled or erased) is not
	// found: order emails rely on it never to write to a closed account.
	StaffGetByID(ctx context.Context, orderID uuid.UUID) (*StaffOrderListRow, error)
	ListByUser(ctx context.Context, params ListOrdersParams) ([]domainorder.Order, error)
	// HasOpen reports whether the customer has an order not yet fulfilled or
	// cancelled — one the bakery still has to make or hand over.
	HasOpen(ctx context.Context, userID uuid.UUID) (bool, error)
	// ListByUserBefore pages a customer's orders newest first by keyset.
	ListByUserBefore(ctx context.Context, userID uuid.UUID, before *PageCursor, limit int32) ([]domainorder.Order, error)
	ListCountByUser(ctx context.Context, userID uuid.UUID, filter OrderHistoryFilter) (int64, error)
	StaffList(ctx context.Context, params StaffListOrdersParams) ([]StaffOrderListRow, error)
	StaffListCount(ctx context.Context, status *domainorder.Status) (int64, error)
	UpdateStatus(ctx context.Context, params UpdateOrderStatusParams) (*domainorder.Order, error)
	CreateItems(ctx context.Context, items []CreateOrderItemParams) error
	ListItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domainorder.Item, error)
	SumItemQuantitiesByOrderIDs(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID]int32, error)
	// ListItemsByOrderIDs and ListStatusEventsByOrderIDs read many orders' lines
	// and histories in one round trip each, keyed by order.
	ListItemsByOrderIDs(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID][]domainorder.Item, error)
	ListStatusEventsByOrderIDs(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID][]domainorder.StatusEvent, error)
	// NextDayNumber hands out the next order number of the bakery day the
	// transaction runs on, and that day. It must run in the transaction that
	// creates the order: the number is held until then.
	NextDayNumber(ctx context.Context) (day time.Time, number int, err error)
	// HoldPickupSlot locks the pickup slot [startsAt, startsAt+length) until
	// the transaction ends and returns how many orders already hold it. It must
	// run in the transaction that creates the order.
	HoldPickupSlot(ctx context.Context, startsAt time.Time, length time.Duration) (int, error)
	// CountCustomerOrdersBetween counts a customer's orders not cancelled with a
	// pickup in [from, to).
	CountCustomerOrdersBetween(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error)
	// CountByPickupTime returns how many orders not cancelled are due at each
	// pickup time in [from, to).
	CountByPickupTime(ctx context.Context, from, to time.Time) ([]PickupCount, error)
	AddStatusEvent(ctx context.Context, params AddOrderStatusEventParams) error
	ListStatusEvents(ctx context.Context, orderID uuid.UUID) ([]domainorder.StatusEvent, error)
}
