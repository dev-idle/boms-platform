package order

import "testing"

func TestCanTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		from Status
		to   Status
		want bool
	}{
		{StatusPending, StatusConfirmed, true},
		{StatusPending, StatusCancelled, true},
		{StatusPending, StatusFulfilled, false},
		{StatusConfirmed, StatusFulfilled, true},
		{StatusConfirmed, StatusCancelled, true},
		{StatusConfirmed, StatusPending, false},
		{StatusFulfilled, StatusCancelled, false},
		{StatusCancelled, StatusConfirmed, false},
		{StatusPending, StatusPending, false},
	}

	for _, tc := range tests {
		if got := CanTransition(tc.from, tc.to); got != tc.want {
			t.Errorf("CanTransition(%q, %q) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}
