package usecase

import (
	"strings"
	"time"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// orderHistoryFilter reads a customer's order-history filter. From and To are
// bakery days and both are included whole: an order placed at 23:30 on To is in.
func orderHistoryFilter(q dto.OrderHistoryQuery) (port.OrderHistoryFilter, error) {
	var filter port.OrderHistoryFilter
	if raw := strings.TrimSpace(q.Status); raw != "" {
		status := domainorder.Status(raw)
		if !status.Valid() {
			return port.OrderHistoryFilter{}, apperrors.ErrValidation.WithDetail("status", "invalid order status")
		}
		filter.Status = &status
	}
	from, err := bakeryDayStart(q.From, "from")
	if err != nil {
		return port.OrderHistoryFilter{}, err
	}
	to, err := bakeryDayStart(q.To, "to")
	if err != nil {
		return port.OrderHistoryFilter{}, err
	}
	if !from.IsZero() && !to.IsZero() && to.Before(from) {
		return port.OrderHistoryFilter{}, apperrors.ErrValidation.WithDetail("to", "must not be before from")
	}
	if !from.IsZero() {
		filter.PlacedFrom = &from
	}
	if !to.IsZero() {
		next := to.AddDate(0, 0, 1)
		filter.PlacedBefore = &next
	}
	return filter, nil
}

// bakeryDayStart returns the instant a bakery day (YYYY-MM-DD) begins, or the
// zero time when value is empty.
func bakeryDayStart(value, field string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	day, err := time.ParseInLocation(domainstore.DayLayout, value, domainstore.Location)
	if err != nil {
		return time.Time{}, apperrors.ErrValidation.WithDetail(field, "use YYYY-MM-DD")
	}
	return day, nil
}
