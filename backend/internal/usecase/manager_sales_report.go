package usecase

import (
	"context"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

const (
	// salesReportMaxDays is the longest range one report covers: a year, leap
	// day included.
	salesReportMaxDays = 366
	// salesReportTopRows is how many items and discount codes a report lists.
	salesReportTopRows int32 = 10
)

// salesGroupings are the periods a report adds up by.
var salesGroupings = map[string]bool{"day": true, "week": true, "month": true}

// ManagerSalesReportUsecase is a manager reading what the bakery sold.
type ManagerSalesReportUsecase struct {
	sales port.SalesReportRepository
}

func NewManagerSalesReportUsecase(sales port.SalesReportRepository) *ManagerSalesReportUsecase {
	return &ManagerSalesReportUsecase{sales: sales}
}

// salesReportParts is what one report reads.
type salesReportParts struct {
	periods    []port.SalesPeriod
	items      []port.SoldItem
	categories []port.CategorySales
	discounts  port.DiscountUse
	codes      []port.DiscountCodeUse
	production port.ProductionTime
}

// Report adds up the bakery days [from, to] (YYYY-MM-DD, no later than
// today, at most a year) by group: the money taken and given back by period,
// and what the orders kept sold — items, categories, discounts — with how
// long orders took to make. Its parts are read at once.
func (u *ManagerSalesReportUsecase) Report(ctx context.Context, from, to, group string, now time.Time) (*dto.SalesReportResponse, error) {
	sales, err := salesRange(from, to, group, now)
	if err != nil {
		return nil, err
	}
	var parts salesReportParts
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		parts.periods, err = u.sales.SalesByPeriod(gctx, sales)
		return err
	})
	g.Go(func() (err error) {
		parts.items, err = u.sales.TopSoldItems(gctx, sales.From, sales.To, salesReportTopRows)
		return err
	})
	g.Go(func() (err error) {
		parts.categories, err = u.sales.SalesByCategory(gctx, sales.From, sales.To)
		return err
	})
	g.Go(func() (err error) {
		parts.discounts, err = u.sales.SalesDiscounts(gctx, sales.From, sales.To)
		return err
	})
	g.Go(func() (err error) {
		parts.codes, err = u.sales.SalesDiscountCodes(gctx, sales.From, sales.To, salesReportTopRows)
		return err
	})
	g.Go(func() (err error) {
		parts.production, err = u.sales.ProductionTime(gctx, sales.From, sales.To)
		return err
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return toSalesReportResponse(sales, parts), nil
}

// toSalesReportResponse maps a report's parts and adds up its totals: net is
// sales less refunds, and the average order the sales over the payments,
// rounded to the nearest cent.
func toSalesReportResponse(sales port.SalesRange, parts salesReportParts) *dto.SalesReportResponse {
	out := &dto.SalesReportResponse{
		From:       sales.FromDay.Format(domainstore.DayLayout),
		To:         sales.ToDay.Format(domainstore.DayLayout),
		Group:      sales.Grouping,
		Periods:    make([]dto.SalesPeriodResponse, 0, len(parts.periods)),
		Items:      make([]dto.SoldItemResponse, 0, len(parts.items)),
		Categories: make([]dto.CategorySalesResponse, 0, len(parts.categories)),
		Discounts: dto.SalesDiscountsResponse{
			Orders:           parts.discounts.Orders,
			DiscountedOrders: parts.discounts.DiscountedOrders,
			DiscountCents:    parts.discounts.DiscountCents,
			Codes:            make([]dto.DiscountCodeUseResponse, 0, len(parts.codes)),
		},
		Production: mapProductionTimeToDTO(parts.production),
	}
	for _, row := range parts.periods {
		out.Periods = append(out.Periods, dto.SalesPeriodResponse{
			Start:        row.Start.Format(domainstore.DayLayout),
			Orders:       row.Orders,
			GrossCents:   row.GrossCents,
			RefundsCents: row.RefundsCents,
			NetCents:     row.GrossCents - row.RefundsCents,
		})
		out.Totals.Orders += row.Orders
		out.Totals.GrossCents += row.GrossCents
		out.Totals.RefundsCents += row.RefundsCents
	}
	out.Totals.NetCents = out.Totals.GrossCents - out.Totals.RefundsCents
	if out.Totals.Orders > 0 {
		average := (out.Totals.GrossCents*2 + out.Totals.Orders) / (out.Totals.Orders * 2)
		out.Totals.AverageOrderCents = &average
	}
	for _, item := range parts.items {
		out.Items = append(out.Items, dto.SoldItemResponse{
			Kind:       string(item.Kind),
			ID:         item.ID.String(),
			Name:       item.Name,
			Quantity:   item.Quantity,
			SalesCents: item.SalesCents,
		})
	}
	for _, category := range parts.categories {
		row := dto.CategorySalesResponse{Name: category.Name, Quantity: category.Quantity, SalesCents: category.SalesCents}
		if category.CategoryID != nil {
			id := category.CategoryID.String()
			row.CategoryID = &id
		}
		out.Categories = append(out.Categories, row)
	}
	for _, code := range parts.codes {
		out.Discounts.Codes = append(out.Discounts.Codes, dto.DiscountCodeUseResponse{
			Code: code.Code, Orders: code.Orders, DiscountCents: code.DiscountCents,
		})
	}
	return out
}

// salesRange checks a report's bakery days and grouping and returns the
// periods to read: [from, to] as days, and as instants from the midnight
// starting from to the one after to.
func salesRange(from, to, group string, now time.Time) (port.SalesRange, error) {
	fromDay, err := time.ParseInLocation(domainstore.DayLayout, strings.TrimSpace(from), domainstore.Location)
	if err != nil {
		return port.SalesRange{}, apperrors.ErrValidation.WithDetail("from", "use YYYY-MM-DD")
	}
	toDay, err := time.ParseInLocation(domainstore.DayLayout, strings.TrimSpace(to), domainstore.Location)
	if err != nil {
		return port.SalesRange{}, apperrors.ErrValidation.WithDetail("to", "use YYYY-MM-DD")
	}
	grouping := strings.TrimSpace(group)
	if !salesGroupings[grouping] {
		return port.SalesRange{}, apperrors.ErrValidation.WithDetail("group", "must be day, week or month")
	}
	switch {
	case domainstore.DayOf(toDay).After(domainstore.DayOf(now)):
		return port.SalesRange{}, apperrors.ErrValidation.WithDetail("to", "must be today or earlier")
	case toDay.Before(fromDay):
		return port.SalesRange{}, apperrors.ErrValidation.WithDetail("to", "must not be before from")
	case !toDay.Before(fromDay.AddDate(0, 0, salesReportMaxDays)):
		return port.SalesRange{}, apperrors.ErrValidation.WithDetail("from", "a report covers at most 366 days")
	}
	return port.SalesRange{
		FromDay:  domainstore.DayOf(fromDay),
		ToDay:    domainstore.DayOf(toDay),
		From:     fromDay,
		To:       toDay.AddDate(0, 0, 1),
		Grouping: grouping,
	}, nil
}
