package order

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
)

// IncidentType is what went wrong with an order: what the system records as it
// happens, or what staff report.
type IncidentType string

const (
	// IncidentBakeryCancelled is the bakery cancelling an order it had accepted.
	IncidentBakeryCancelled IncidentType = "bakery_cancelled"
	// IncidentReadyLate is an order ready only after its pickup time.
	IncidentReadyLate IncidentType = "ready_late"
	// IncidentNoShow is an order not collected on its pickup day.
	IncidentNoShow IncidentType = "no_show"
	// IncidentPaymentFailed is the provider refusing an order's payment.
	IncidentPaymentFailed IncidentType = "payment_failed"
	// IncidentPaymentExpired is an order not paid in time.
	IncidentPaymentExpired IncidentType = "payment_expired"
	// IncidentRefunded is an order's payment going back to its buyer.
	IncidentRefunded IncidentType = "refunded"
	// IncidentPaymentAnomaly flags a customer whose payments failed or expired
	// PaymentAnomalyThreshold times within PaymentAnomalyWindow.
	IncidentPaymentAnomaly IncidentType = "payment_anomaly"
	// IncidentWrongItems is staff reporting an order made with the wrong items.
	IncidentWrongItems IncidentType = "wrong_items"
	// IncidentCustomMismatch is staff reporting a custom cake made unlike the
	// customer asked.
	IncidentCustomMismatch IncidentType = "custom_mismatch"
	// IncidentOther is staff reporting any other problem, which the note says.
	IncidentOther IncidentType = "other"
)

const (
	// PaymentAnomalyWindow is how far back a customer's failed and expired
	// payments count towards a flag; one flag covers them.
	PaymentAnomalyWindow = 24 * time.Hour
	// PaymentAnomalyThreshold is how many failed or expired payments within the
	// window flag a customer: one is a mistake, three are a pattern.
	PaymentAnomalyThreshold = 3
	// maxIncidentNoteLength is the most characters a note holds.
	maxIncidentNoteLength = 500
)

// IncidentSource is who records an incident.
type IncidentSource string

const (
	IncidentSourceAuto   IncidentSource = "auto"
	IncidentSourceManual IncidentSource = "manual"
)

// Valid reports whether t is a type this package defines.
func (t IncidentType) Valid() bool {
	switch t {
	case IncidentBakeryCancelled, IncidentReadyLate, IncidentNoShow, IncidentPaymentFailed, IncidentPaymentExpired,
		IncidentRefunded, IncidentPaymentAnomaly, IncidentWrongItems, IncidentCustomMismatch, IncidentOther:
		return true
	default:
		return false
	}
}

// Source is who records an incident of this type: staff report the wrong
// items, a custom cake made unlike asked and any other problem; the system
// records every other as it happens.
func (t IncidentType) Source() IncidentSource {
	switch t {
	case IncidentWrongItems, IncidentCustomMismatch, IncidentOther:
		return IncidentSourceManual
	default:
		return IncidentSourceAuto
	}
}

// PaymentProblem reports whether an incident of this type counts towards a
// payment anomaly flag.
func (t IncidentType) PaymentProblem() bool {
	return t == IncidentPaymentFailed || t == IncidentPaymentExpired
}

// NewReportedIncident checks what staff report about an order: a type staff
// report, and a note of 1 to 500 plain characters that may break across lines,
// returned trimmed.
func NewReportedIncident(incidentType, note string) (IncidentType, string, error) {
	t := IncidentType(incidentType)
	if t.Source() != IncidentSourceManual {
		return "", "", ErrInvalidIncidentType
	}
	note = domainstore.TrimLines(note)
	if note == "" || utf8.RuneCountInString(note) > maxIncidentNoteLength || !domainstore.PlainLines(note) {
		return "", "", ErrInvalidIncidentNote
	}
	return t, note, nil
}

// MoveIncident is the incident an order's move from one status to another
// records, if any: staff cancelling an order the bakery had accepted, a missed
// pickup, or an unpaid order expiring. byStaff reports whether staff made the
// move.
func MoveIncident(from, to Status, byStaff bool) (IncidentType, bool) {
	switch {
	case to == StatusCancelled && byStaff && (from == StatusConfirmed || from == StatusInProduction || from == StatusReady):
		return IncidentBakeryCancelled, true
	case to == StatusNoShow:
		return IncidentNoShow, true
	case to == StatusExpired:
		return IncidentPaymentExpired, true
	default:
		return "", false
	}
}

// ReadyLate reports whether the order became ready only after its pickup time,
// as of now.
func (o Order) ReadyLate(now time.Time) bool {
	return o.Status == StatusReady && o.PickupAt != nil && now.After(*o.PickupAt)
}

// Incident is something that went wrong with an order. Note is what staff
// wrote reporting it; nil for an incident the system recorded.
type Incident struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	Type      IncidentType
	Note      *string
	CreatedAt time.Time
}

// TopicIncidentRecorded announces an incident recorded with an order.
const TopicIncidentRecorded domainevent.Topic = "incident.recorded"

// IncidentRecordedEvent tells managers, who follow the incident log, that one
// was recorded.
func IncidentRecordedEvent(incident Incident) domainevent.Event {
	return domainevent.New(TopicIncidentRecorded, domainevent.Audience{
		Roles: []domainuser.Role{domainuser.RoleManager},
	}, map[string]string{
		"incident_id": incident.ID.String(),
		"type":        string(incident.Type),
	})
}
