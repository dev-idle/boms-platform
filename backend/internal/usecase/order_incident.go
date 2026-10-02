package usecase

import (
	"context"

	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

// recordIncident records an incident with an order in the transaction that
// found it, tells managers and returns it. A failed or expired payment that is
// one too many for its customer also flags them.
func recordIncident(
	txCtx context.Context,
	incidents port.OrderIncidentRepository,
	events port.EventOutbox,
	params port.AddOrderIncidentParams,
) (*domainorder.Incident, error) {
	incident, err := incidents.AddIncident(txCtx, params)
	if err != nil {
		return nil, err
	}
	if err := events.Add(txCtx, domainorder.IncidentRecordedEvent(*incident)); err != nil {
		return nil, err
	}
	if !params.Type.PaymentProblem() {
		return incident, nil
	}
	flag, err := incidents.FlagPaymentAnomaly(txCtx, params.OrderID,
		domainorder.PaymentAnomalyWindow, domainorder.PaymentAnomalyThreshold)
	if err != nil || flag == nil {
		return incident, err
	}
	return incident, events.Add(txCtx, domainorder.IncidentRecordedEvent(*flag))
}
