package order

import (
	"slices"
	"strings"
	"unicode/utf8"

	domainstore "github.com/boms/backend/internal/domain/store"
)

// maxCancelReasonLength is the longest reason a customer is shown.
const maxCancelReasonLength = 200

// NewCancelReason checks the reason the bakery gives a customer for cancelling
// their order and returns it trimmed.
func NewCancelReason(raw string) (string, error) {
	reason := strings.TrimSpace(raw)
	if reason == "" || utf8.RuneCountInString(reason) > maxCancelReasonLength || !domainstore.PlainText(reason) {
		return "", ErrInvalidCancelReason
	}
	return reason, nil
}

// VisibleToBaker reports whether an order in this status is production work the
// kitchen may see: accepted, being made, or waiting at the counter.
func (s Status) VisibleToBaker() bool {
	switch s {
	case StatusConfirmed, StatusInProduction, StatusReady:
		return true
	default:
		return false
	}
}

// VisibleToStaff reports whether an order in this status is the bakery's to
// see: one not paid, now or ever, is the customer's alone.
func (s Status) VisibleToStaff() bool {
	return s != StatusAwaitingPayment && s != StatusExpired
}

// SeenByStaff reports whether an order now in status, with this history, is
// the bakery's to see: one cancelled is only if it was accepted before, so an
// order its customer cancelled before paying stays theirs alone.
func SeenByStaff(status Status, history []StatusEvent) bool {
	if status != StatusCancelled {
		return status.VisibleToStaff()
	}
	return slices.ContainsFunc(history, func(e StatusEvent) bool {
		return e.To == StatusPending || e.To == StatusConfirmed
	})
}

// CanStaffTransition reports whether staff may change order status at the counter.
func CanStaffTransition(from, to Status) bool {
	if from == to {
		return false
	}
	if !from.Valid() || !to.Valid() {
		return false
	}
	switch from {
	case StatusPending:
		return to == StatusConfirmed || to == StatusCancelled
	case StatusConfirmed, StatusInProduction:
		return to == StatusCancelled
	case StatusReady:
		return to == StatusFulfilled || to == StatusCancelled
	default:
		return false
	}
}
