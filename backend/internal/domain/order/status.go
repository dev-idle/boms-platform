package order

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
