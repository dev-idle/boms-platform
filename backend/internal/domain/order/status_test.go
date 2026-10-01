package order

import (
	"errors"
	"strings"
	"testing"
)

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

func TestPaidStatus(t *testing.T) {
	t.Parallel()
	if got := PaidStatus(true); got != StatusPending {
		t.Errorf("a custom order waits for staff review, got %q", got)
	}
	if got := PaidStatus(false); got != StatusConfirmed {
		t.Errorf("any other goes to the bakery, got %q", got)
	}
}

func TestStatus_BeforeProduction(t *testing.T) {
	t.Parallel()
	for _, s := range []Status{StatusAwaitingPayment, StatusPending, StatusConfirmed} {
		if !s.BeforeProduction() {
			t.Errorf("%q: the customer may still cancel or move it", s)
		}
	}
	for _, s := range []Status{StatusInProduction, StatusReady, StatusFulfilled, StatusCancelled, StatusExpired, StatusNoShow} {
		if s.BeforeProduction() {
			t.Errorf("%q: being made or closed", s)
		}
	}
}

func TestNewCancelReason(t *testing.T) {
	t.Parallel()
	got, err := NewCancelReason("  Out of matcha today  ")
	if err != nil || got != "Out of matcha today" {
		t.Fatalf("got %q, %v; want the trimmed reason", got, err)
	}
	for name, raw := range map[string]string{
		"empty":          "   ",
		"too_long":       strings.Repeat("a", 201),
		"control":        "Out of\nmatcha",
		"direction_mark": "Out of " + string(rune(0x202e)) + "matcha",
	} {
		if _, err := NewCancelReason(raw); !errors.Is(err, ErrInvalidCancelReason) {
			t.Errorf("%s: got %v, want ErrInvalidCancelReason", name, err)
		}
	}
}
