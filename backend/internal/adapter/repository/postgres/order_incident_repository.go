package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// customerIncidentLockNamespace keeps the lock on a customer's flags apart from
// any other advisory lock.
const customerIncidentLockNamespace = 1103

func (r *OrderRepository) AddIncident(ctx context.Context, params port.AddOrderIncidentParams) (*domainorder.Incident, error) {
	var actorID *uuid.UUID
	var actorRole *sqlcgen.UserRole
	if params.Actor != nil {
		role, err := toSQLRole(params.Actor.Role)
		if err != nil {
			return nil, err
		}
		actorID, actorRole = &params.Actor.ID, &role
	}
	row, err := r.q(ctx).CreateOrderIncident(ctx, sqlcgen.CreateOrderIncidentParams{
		OrderID:   params.OrderID,
		Type:      sqlcgen.OrderIncidentType(params.Type),
		Note:      params.Note,
		ActorID:   actorID,
		ActorRole: actorRole,
	})
	if err != nil {
		return nil, mapRepoError(err, "create order incident")
	}
	incident := incidentFromRow(row)
	return &incident, nil
}

func (r *OrderRepository) FlagPaymentAnomaly(
	ctx context.Context,
	orderID uuid.UUID,
	window time.Duration,
	threshold int32,
) (*domainorder.Incident, error) {
	if txFromContext(ctx) == nil {
		return nil, apperrors.Errorf("flag payment anomaly: requires a transaction")
	}
	// The count runs as a statement of its own after the lock, so it sees what
	// a transaction that held the lock before committed.
	if err := r.q(ctx).LockCustomerIncidents(ctx, sqlcgen.LockCustomerIncidentsParams{
		Namespace: customerIncidentLockNamespace,
		OrderID:   orderID,
	}); err != nil {
		return nil, mapRepoError(err, "lock customer incidents")
	}
	row, err := r.q(ctx).FlagPaymentAnomaly(ctx, sqlcgen.FlagPaymentAnomalyParams{
		OrderID:       orderID,
		Threshold:     threshold,
		WindowSeconds: window.Seconds(),
	})
	if err != nil {
		err = mapRepoError(err, "flag payment anomaly")
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	incident := incidentFromRow(row)
	return &incident, nil
}

func (r *OrderRepository) ListIncidents(ctx context.Context, params port.ListOrderIncidentsParams) ([]port.ManagerOrderIncident, error) {
	rows, err := r.q(ctx).ListOrderIncidents(ctx, sqlcgen.ListOrderIncidentsParams{
		FromAt: params.From,
		ToAt:   params.To,
		Type:   optionalIncidentType(params.Type),
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "list order incidents")
	}
	out := make([]port.ManagerOrderIncident, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.ManagerOrderIncident{
			Incident: domainorder.Incident{
				ID:        row.ID,
				OrderID:   row.OrderID,
				Type:      domainorder.IncidentType(row.Type),
				Note:      row.Note,
				CreatedAt: row.CreatedAt,
			},
			OrderCode:    row.OrderCode,
			PickupAt:     row.PickupAt,
			ActorName:    row.ActorName,
			CancelReason: row.CancelReason,
		})
	}
	return out, nil
}

func (r *OrderRepository) CountIncidents(
	ctx context.Context,
	from, to time.Time,
	incidentType *domainorder.IncidentType,
) (int64, error) {
	count, err := r.q(ctx).CountOrderIncidents(ctx, sqlcgen.CountOrderIncidentsParams{
		FromAt: from,
		ToAt:   to,
		Type:   optionalIncidentType(incidentType),
	})
	if err != nil {
		return 0, mapRepoError(err, "count order incidents")
	}
	return count, nil
}

func (r *OrderRepository) CountIncidentsByType(ctx context.Context, from, to time.Time) ([]port.IncidentTypeCount, error) {
	rows, err := r.q(ctx).CountOrderIncidentsByType(ctx, sqlcgen.CountOrderIncidentsByTypeParams{FromAt: from, ToAt: to})
	if err != nil {
		return nil, mapRepoError(err, "count order incidents by type")
	}
	out := make([]port.IncidentTypeCount, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.IncidentTypeCount{Type: domainorder.IncidentType(row.Type), Count: row.Count})
	}
	return out, nil
}

func (r *OrderRepository) ListIncidentsByOrderIDs(
	ctx context.Context,
	orderIDs []uuid.UUID,
) (map[uuid.UUID][]domainorder.Incident, error) {
	rows, err := r.q(ctx).ListOrderIncidentsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, mapRepoError(err, "list order incidents by order ids")
	}
	out := make(map[uuid.UUID][]domainorder.Incident, len(orderIDs))
	for _, row := range rows {
		out[row.OrderID] = append(out[row.OrderID], incidentFromRow(row))
	}
	return out, nil
}

func (r *OrderRepository) EraseIncidentNotes(ctx context.Context, userID uuid.UUID) error {
	if err := r.q(ctx).EraseOrderIncidentNotes(ctx, userID); err != nil {
		return mapRepoError(err, "erase order incident notes")
	}
	return nil
}

func incidentFromRow(row sqlcgen.OrderIncident) domainorder.Incident {
	return domainorder.Incident{
		ID:        row.ID,
		OrderID:   row.OrderID,
		Type:      domainorder.IncidentType(row.Type),
		Note:      row.Note,
		CreatedAt: row.CreatedAt,
	}
}

func optionalIncidentType(t *domainorder.IncidentType) *sqlcgen.OrderIncidentType {
	if t == nil {
		return nil
	}
	mapped := sqlcgen.OrderIncidentType(*t)
	return &mapped
}
