package order

// CanTransition reports whether a staff-operated status change is allowed.
func CanTransition(from, to Status) bool {
	if from == to {
		return false
	}
	if !from.Valid() || !to.Valid() {
		return false
	}
	switch from {
	case StatusPending:
		return to == StatusConfirmed || to == StatusCancelled
	case StatusConfirmed:
		return to == StatusFulfilled || to == StatusCancelled
	default:
		return false
	}
}
