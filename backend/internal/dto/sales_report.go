package dto

// SalesReportResponse is what the bakery sold over the bakery days [From, To]
// (YYYY-MM-DD), by Group (day, week or month). Totals and Periods count the
// money: payments taken and refunds made in the range. Items, Categories and
// Discounts cover the orders kept: paid in the range and not cancelled.
// Production covers the orders that became ready in the range.
type SalesReportResponse struct {
	From       string                  `json:"from"`
	To         string                  `json:"to"`
	Group      string                  `json:"group"`
	Totals     SalesTotalsResponse     `json:"totals"`
	Periods    []SalesPeriodResponse   `json:"periods"`
	Items      []SoldItemResponse      `json:"items"`
	Categories []CategorySalesResponse `json:"categories"`
	Discounts  SalesDiscountsResponse  `json:"discounts"`
	Production ProductionTimeResponse  `json:"production"`
}

// SalesTotalsResponse is the range's payments and refunds; the average order
// is the payments taken over how many, null when there were none.
type SalesTotalsResponse struct {
	Orders            int64  `json:"orders"`
	GrossCents        int64  `json:"gross_cents"`
	RefundsCents      int64  `json:"refunds_cents"`
	NetCents          int64  `json:"net_cents"`
	AverageOrderCents *int64 `json:"average_order_cents"`
}

// SalesPeriodResponse is one period, from its first bakery day in the range.
type SalesPeriodResponse struct {
	Start        string `json:"start"`
	Orders       int64  `json:"orders"`
	GrossCents   int64  `json:"gross_cents"`
	RefundsCents int64  `json:"refunds_cents"`
	NetCents     int64  `json:"net_cents"`
}

// SoldItemResponse is a product or a combo and what the kept orders sold of
// it, before any order discount.
type SoldItemResponse struct {
	Kind       string `json:"kind"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	Quantity   int64  `json:"quantity"`
	SalesCents int64  `json:"sales_cents"`
}

// CategorySalesResponse is what the kept orders sold in one category; null
// for the combos, which have none.
type CategorySalesResponse struct {
	CategoryID *string `json:"category_id"`
	Name       *string `json:"name"`
	Quantity   int64   `json:"quantity"`
	SalesCents int64   `json:"sales_cents"`
}

// SalesDiscountsResponse is how many orders were kept, how many used a
// discount code, how much the codes took off, and each code's share.
type SalesDiscountsResponse struct {
	Orders           int64                     `json:"orders"`
	DiscountedOrders int64                     `json:"discounted_orders"`
	DiscountCents    int64                     `json:"discount_cents"`
	Codes            []DiscountCodeUseResponse `json:"codes"`
}

// DiscountCodeUseResponse is one code's kept orders and what it took off.
type DiscountCodeUseResponse struct {
	Code          string `json:"code"`
	Orders        int64  `json:"orders"`
	DiscountCents int64  `json:"discount_cents"`
}

// ProductionTimeResponse is how many orders became ready and their average
// minutes from the first ticket started to ready, null when none did.
type ProductionTimeResponse struct {
	Orders         int64  `json:"orders"`
	AverageMinutes *int32 `json:"average_minutes"`
}
