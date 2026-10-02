package usecase

import (
	"context"
	"strings"
	"time"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// ManagerIncidentUsecase is a manager reading what went wrong with orders, a
// bakery week at a time.
type ManagerIncidentUsecase struct {
	incidents port.OrderIncidentRepository
}

func NewManagerIncidentUsecase(incidents port.OrderIncidentRepository) *ManagerIncidentUsecase {
	return &ManagerIncidentUsecase{incidents: incidents}
}

// List returns a page of the incidents recorded in the week starting on the
// Monday week names, latest first. typeFilter narrows it to one type.
func (u *ManagerIncidentUsecase) List(
	ctx context.Context,
	week, typeFilter string,
	page, pageSize int32,
) ([]dto.ManagerOrderIncidentResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)
	from, to, err := incidentWeek(week)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	var incidentType *domainorder.IncidentType
	if trimmed := strings.TrimSpace(typeFilter); trimmed != "" {
		parsed := domainorder.IncidentType(trimmed)
		if !parsed.Valid() {
			return nil, 0, page, pageSize, apperrors.ErrValidation.WithDetail("type", "unknown incident type")
		}
		incidentType = &parsed
	}
	rows, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.ManagerOrderIncident, error) {
			return u.incidents.ListIncidents(ctx, port.ListOrderIncidentsParams{
				From:   from,
				To:     to,
				Type:   incidentType,
				Limit:  pageSize,
				Offset: utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.incidents.CountIncidents(ctx, from, to, incidentType)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.ManagerOrderIncidentResponse, 0, len(rows))
	for _, row := range rows {
		if row.CancelReason != nil {
			row.Incident.Note = row.CancelReason
		}
		out = append(out, dto.ManagerOrderIncidentResponse{
			OrderIncidentResponse: toOrderIncidentResponse(row.Incident),
			OrderCode:             row.OrderCode,
			PickupAt:              row.PickupAt,
			ActorName:             row.ActorName,
		})
	}
	return out, total, page, pageSize, nil
}

// Summary counts the incidents of each type recorded in the week starting on
// the Monday week names.
func (u *ManagerIncidentUsecase) Summary(ctx context.Context, week string) (*dto.IncidentSummaryResponse, error) {
	from, to, err := incidentWeek(week)
	if err != nil {
		return nil, err
	}
	counts, err := u.incidents.CountIncidentsByType(ctx, from, to)
	if err != nil {
		return nil, err
	}
	out := &dto.IncidentSummaryResponse{Types: make([]dto.IncidentTypeCountResponse, 0, len(counts))}
	for _, count := range counts {
		out.Types = append(out.Types, dto.IncidentTypeCountResponse{Type: string(count.Type), Count: count.Count})
	}
	return out, nil
}

// incidentWeek is the bakery week [from, to) starting on the Monday week
// names, as YYYY-MM-DD.
func incidentWeek(week string) (time.Time, time.Time, error) {
	from, err := time.ParseInLocation(domainstore.DayLayout, strings.TrimSpace(week), domainstore.Location)
	if err != nil || from.Weekday() != time.Monday {
		return time.Time{}, time.Time{}, apperrors.ErrValidation.WithDetail("week", "use the Monday the week starts on, as YYYY-MM-DD")
	}
	return from, from.AddDate(0, 0, 7), nil
}
