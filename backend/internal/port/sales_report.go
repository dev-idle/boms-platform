package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domaincart "github.com/boms/backend/internal/domain/cart"
)

// SalesRange asks for the payments taken and refunds made in [From, To),
// by Grouping (day, week or month), every period of the bakery days
// [FromDay, ToDay] listed.
type SalesRange struct {
	FromDay  time.Time
	ToDay    time.Time
	From     time.Time
	To       time.Time
	Grouping string
}

// SalesPeriod is what one period took: how many payments and how much, and
// how much went back. Start is its first bakery day in the range.
type SalesPeriod struct {
	Start        time.Time
	Orders       int64
	GrossCents   int64
	RefundsCents int64
}

// SoldItem is a product or a combo and what the orders kept sold of it.
type SoldItem struct {
	Kind       domaincart.LineType
	ID         uuid.UUID
	Name       string
	Quantity   int64
	SalesCents int64
}

// CategorySales is what the orders kept sold in one category; a nil category
// is the combos, which have none.
type CategorySales struct {
	CategoryID *uuid.UUID
	Name       *string
	Quantity   int64
	SalesCents int64
}

// DiscountUse is how many orders were kept, how many used a discount code,
// and how much the codes took off.
type DiscountUse struct {
	Orders           int64
	DiscountedOrders int64
	DiscountCents    int64
}

// DiscountCodeUse is how many kept orders used one code and what it took off.
type DiscountCodeUse struct {
	Code          string
	Orders        int64
	DiscountCents int64
}

// ProductionTime is how many orders became ready and their average minutes
// from the first ticket started to ready, 0 when none did.
type ProductionTime struct {
	Orders         int64
	AverageMinutes int32
}

// SalesReportRepository reads what the bakery sold. The breakdowns cover the
// orders kept: paid in [from, to) and not cancelled.
type SalesReportRepository interface {
	SalesByPeriod(ctx context.Context, params SalesRange) ([]SalesPeriod, error)
	TopSoldItems(ctx context.Context, from, to time.Time, limit int32) ([]SoldItem, error)
	SalesByCategory(ctx context.Context, from, to time.Time) ([]CategorySales, error)
	SalesDiscounts(ctx context.Context, from, to time.Time) (DiscountUse, error)
	SalesDiscountCodes(ctx context.Context, from, to time.Time, limit int32) ([]DiscountCodeUse, error)
	// ProductionTime covers the orders that became ready in [from, to).
	ProductionTime(ctx context.Context, from, to time.Time) (ProductionTime, error)
}
