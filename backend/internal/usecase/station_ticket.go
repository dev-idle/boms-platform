package usecase

import (
	"context"
	"strings"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// listStationTickets is a page of a station's queue, soonest pickup first,
// narrowed to one ticket status when statusFilter names one.
func listStationTickets(
	ctx context.Context,
	tickets port.TicketRepository,
	station domaincategory.Station,
	page, pageSize int32,
	statusFilter string,
) ([]dto.StationTicketResponse, int64, int32, int32, error) {
	page, pageSize = normalizeOrderListPage(page, pageSize)
	var status *domainorder.TicketStatus
	if trimmed := strings.TrimSpace(statusFilter); trimmed != "" {
		parsed := domainorder.TicketStatus(trimmed)
		if !parsed.Valid() || parsed == domainorder.TicketCancelled {
			return nil, 0, page, pageSize, apperrors.ErrValidation.WithDetail("status", "queued, in_progress or ready")
		}
		status = &parsed
	}
	rows, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.StationTicket, error) {
			return tickets.ListByStation(ctx, port.StationTicketsParams{
				Station: station,
				Status:  status,
				Limit:   pageSize,
				Offset:  utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return tickets.CountByStation(ctx, station, status)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.StationTicketResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toStationTicketResponse(row))
	}
	return out, total, page, pageSize, nil
}
