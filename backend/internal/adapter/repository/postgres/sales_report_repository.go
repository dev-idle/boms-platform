package postgres

import (
	"context"
	"time"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaincart "github.com/boms/backend/internal/domain/cart"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
)

// SalesReportRepository reads what the bakery sold.
type SalesReportRepository struct {
	queries *sqlcgen.Queries
}

func NewSalesReportRepository(pool *Pool) *SalesReportRepository {
	return &SalesReportRepository{queries: pool.Queries()}
}

var _ port.SalesReportRepository = (*SalesReportRepository)(nil)

func (r *SalesReportRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// SalesByPeriod implements port.SalesReportRepository.
func (r *SalesReportRepository) SalesByPeriod(ctx context.Context, params port.SalesRange) ([]port.SalesPeriod, error) {
	rows, err := r.q(ctx).SalesByPeriod(ctx, sqlcgen.SalesByPeriodParams{
		FromDay:  params.FromDay,
		ToDay:    params.ToDay,
		Grouping: params.Grouping,
		Zone:     domainstore.Location.String(),
		FromAt:   params.From,
		ToAt:     params.To,
	})
	if err != nil {
		return nil, mapRepoError(err, "sales by period")
	}
	out := make([]port.SalesPeriod, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.SalesPeriod{
			Start:        row.Start,
			Orders:       row.Orders,
			GrossCents:   row.GrossCents,
			RefundsCents: row.RefundsCents,
		})
	}
	return out, nil
}

// TopSoldItems implements port.SalesReportRepository.
func (r *SalesReportRepository) TopSoldItems(ctx context.Context, from, to time.Time, limit int32) ([]port.SoldItem, error) {
	rows, err := r.q(ctx).TopSoldItems(ctx, sqlcgen.TopSoldItemsParams{FromAt: from, ToAt: to, MaxRows: limit})
	if err != nil {
		return nil, mapRepoError(err, "top sold items")
	}
	out := make([]port.SoldItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.SoldItem{
			Kind:       domaincart.LineType(row.LineType),
			ID:         row.ItemID,
			Name:       row.Name,
			Quantity:   row.Quantity,
			SalesCents: row.SalesCents,
		})
	}
	return out, nil
}

// SalesByCategory implements port.SalesReportRepository.
func (r *SalesReportRepository) SalesByCategory(ctx context.Context, from, to time.Time) ([]port.CategorySales, error) {
	rows, err := r.q(ctx).SalesByCategory(ctx, sqlcgen.SalesByCategoryParams{FromAt: from, ToAt: to})
	if err != nil {
		return nil, mapRepoError(err, "sales by category")
	}
	out := make([]port.CategorySales, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.CategorySales{
			CategoryID: row.CategoryID,
			Name:       row.CategoryName,
			Quantity:   row.Quantity,
			SalesCents: row.SalesCents,
		})
	}
	return out, nil
}

// SalesDiscounts implements port.SalesReportRepository.
func (r *SalesReportRepository) SalesDiscounts(ctx context.Context, from, to time.Time) (port.DiscountUse, error) {
	row, err := r.q(ctx).SalesDiscounts(ctx, sqlcgen.SalesDiscountsParams{FromAt: from, ToAt: to})
	if err != nil {
		return port.DiscountUse{}, mapRepoError(err, "sales discounts")
	}
	return port.DiscountUse{
		Orders:           row.Orders,
		DiscountedOrders: row.DiscountedOrders,
		DiscountCents:    row.DiscountCents,
	}, nil
}

// SalesDiscountCodes implements port.SalesReportRepository.
func (r *SalesReportRepository) SalesDiscountCodes(ctx context.Context, from, to time.Time, limit int32) ([]port.DiscountCodeUse, error) {
	rows, err := r.q(ctx).SalesDiscountCodes(ctx, sqlcgen.SalesDiscountCodesParams{FromAt: from, ToAt: to, MaxRows: limit})
	if err != nil {
		return nil, mapRepoError(err, "sales discount codes")
	}
	out := make([]port.DiscountCodeUse, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.DiscountCodeUse{Code: row.Code, Orders: row.Orders, DiscountCents: row.DiscountCents})
	}
	return out, nil
}

// ProductionTime implements port.SalesReportRepository.
func (r *SalesReportRepository) ProductionTime(ctx context.Context, from, to time.Time) (port.ProductionTime, error) {
	row, err := r.q(ctx).ProductionTime(ctx, sqlcgen.ProductionTimeParams{FromAt: from, ToAt: to})
	if err != nil {
		return port.ProductionTime{}, mapRepoError(err, "production time")
	}
	return port.ProductionTime{Orders: row.Orders, AverageMinutes: row.AverageMinutes}, nil
}
