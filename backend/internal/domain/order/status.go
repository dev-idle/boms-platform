package order

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

// CanBakerTransition reports whether kitchen staff may advance production status.
func CanBakerTransition(from, to Status) bool {
	if from == to {
		return false
	}
	if !from.Valid() || !to.Valid() {
		return false
	}
	switch from {
	case StatusConfirmed:
		return to == StatusInProduction
	case StatusInProduction:
		return to == StatusReady
	default:
		return false
	}
}
