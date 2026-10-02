package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// ReportIncident records what staff found wrong with an order they see — the
// wrong items, a custom cake made unlike asked, or another problem their note
// explains — and tells managers.
func (u *StaffOrderUsecase) ReportIncident(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	orderID uuid.UUID,
	req dto.ReportOrderIncidentRequest,
) (*dto.OrderIncidentResponse, error) {
	incidentType, note, err := domainorder.NewReportedIncident(req.Type, req.Note)
	if err != nil {
		return nil, err
	}
	row, err := u.orders.StaffGetByID(ctx, orderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, domainorder.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	seen, err := seenByStaff(ctx, u.orders, row.Order)
	if err != nil {
		return nil, err
	}
	if !seen {
		return nil, domainorder.ErrNotFound
	}
	var incident *domainorder.Incident
	err = u.tx.WithTx(ctx, func(txCtx context.Context) (err error) {
		// What staff write is about the customer: their account is held, as a
		// message about the order holds it, so an erasure waits and clears it.
		if row.Order.UserID != nil {
			if err := holdCustomer(txCtx, u.users, *row.Order.UserID); err != nil {
				return err
			}
		}
		incident, err = recordIncident(txCtx, u.orders, u.events, port.AddOrderIncidentParams{
			OrderID: orderID,
			Type:    incidentType,
			Note:    &note,
			Actor:   &port.OrderActor{ID: actorID, Role: actorRole},
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	out := toOrderIncidentResponse(*incident)
	return &out, nil
}
