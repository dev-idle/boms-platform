package order

import "testing"

func TestCanStaffTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		from Status
		to   Status
		want bool
	}{
		{StatusPending, StatusConfirmed, true},
		{StatusPending, StatusCancelled, true},
		{StatusPending, StatusFulfilled, false},
		{StatusPending, StatusInProduction, false},
		{StatusConfirmed, StatusCancelled, true},
		{StatusConfirmed, StatusInProduction, false},
		{StatusConfirmed, StatusFulfilled, false},
		{StatusInProduction, StatusCancelled, true},
		{StatusInProduction, StatusReady, false},
		{StatusReady, StatusFulfilled, true},
		{StatusReady, StatusCancelled, true},
		{StatusReady, StatusInProduction, false},
		{StatusFulfilled, StatusCancelled, false},
		{StatusCancelled, StatusConfirmed, false},
	}

	for _, tc := range tests {
		if got := CanStaffTransition(tc.from, tc.to); got != tc.want {
			t.Errorf("CanStaffTransition(%q, %q) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestCanBakerTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		from Status
		to   Status
		want bool
	}{
		{StatusConfirmed, StatusInProduction, true},
		{StatusConfirmed, StatusReady, false},
		{StatusConfirmed, StatusCancelled, false},
		{StatusInProduction, StatusReady, true},
		{StatusInProduction, StatusFulfilled, false},
		{StatusPending, StatusInProduction, false},
		{StatusReady, StatusFulfilled, false},
	}

	for _, tc := range tests {
		if got := CanBakerTransition(tc.from, tc.to); got != tc.want {
			t.Errorf("CanBakerTransition(%q, %q) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestStatusValid(t *testing.T) {
	t.Parallel()

	valid := []Status{
		StatusPending,
		StatusConfirmed,
		StatusInProduction,
		StatusReady,
		StatusFulfilled,
		StatusCancelled,
	}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("Status(%q).Valid() = false, want true", s)
		}
	}
	if Status("bogus").Valid() {
		t.Fatal("Status(bogus).Valid() = true, want false")
	}
}
