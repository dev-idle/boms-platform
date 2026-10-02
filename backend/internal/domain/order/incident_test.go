package order

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveIncident(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		from    Status
		to      Status
		byStaff bool
		want    IncidentType
	}{
		{name: "staff_cancel_an_accepted_order", from: StatusConfirmed, to: StatusCancelled, byStaff: true, want: IncidentBakeryCancelled},
		{name: "staff_cancel_an_order_being_made", from: StatusInProduction, to: StatusCancelled, byStaff: true, want: IncidentBakeryCancelled},
		{name: "staff_cancel_a_ready_order", from: StatusReady, to: StatusCancelled, byStaff: true, want: IncidentBakeryCancelled},
		{name: "staff_decline_a_custom_request", from: StatusPending, to: StatusCancelled, byStaff: true},
		{name: "a_customer_cancels", from: StatusConfirmed, to: StatusCancelled},
		{name: "a_pickup_is_missed", from: StatusReady, to: StatusNoShow, want: IncidentNoShow},
		{name: "an_unpaid_order_expires", from: StatusAwaitingPayment, to: StatusExpired, want: IncidentPaymentExpired},
		{name: "an_order_is_collected", from: StatusReady, to: StatusFulfilled, byStaff: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := MoveIncident(tc.from, tc.to, tc.byStaff)
			assert.Equal(t, tc.want != "", ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewReportedIncident(t *testing.T) {
	t.Parallel()

	t.Run("takes_a_type_staff_report_and_trims_the_note", func(t *testing.T) {
		t.Parallel()
		got, note, err := NewReportedIncident("custom_mismatch", "  Wrote Mai instead of May.\r\nCustomer took it.  ")
		require.NoError(t, err)
		assert.Equal(t, IncidentCustomMismatch, got)
		assert.Equal(t, "Wrote Mai instead of May.\nCustomer took it.", note)
	})

	t.Run("refuses_a_type_the_system_records", func(t *testing.T) {
		t.Parallel()
		for _, raw := range []string{"no_show", "payment_anomaly", "late", ""} {
			_, _, err := NewReportedIncident(raw, "Note")
			assert.ErrorIs(t, err, ErrInvalidIncidentType, raw)
		}
	})

	t.Run("refuses_an_empty_long_or_hidden_note", func(t *testing.T) {
		t.Parallel()
		for _, raw := range []string{" \n ", strings.Repeat("a", 501), "Box" + string(rune(0x202e)) + "damaged"} {
			_, _, err := NewReportedIncident("other", raw)
			assert.ErrorIs(t, err, ErrInvalidIncidentNote)
		}
		_, note, err := NewReportedIncident("other", strings.Repeat("é", 500))
		require.NoError(t, err, "500 characters, however many bytes")
		assert.Len(t, []rune(note), 500)
	})
}

func TestIncidentType_Source(t *testing.T) {
	t.Parallel()
	assert.Equal(t, IncidentSourceManual, IncidentWrongItems.Source())
	assert.Equal(t, IncidentSourceManual, IncidentOther.Source())
	assert.Equal(t, IncidentSourceAuto, IncidentReadyLate.Source())
	assert.Equal(t, IncidentSourceAuto, IncidentPaymentAnomaly.Source())
	assert.True(t, IncidentPaymentFailed.PaymentProblem())
	assert.True(t, IncidentPaymentExpired.PaymentProblem())
	assert.False(t, IncidentPaymentAnomaly.PaymentProblem(), "a flag does not count towards another")
}

func TestOrder_ReadyLate(t *testing.T) {
	t.Parallel()
	pickup := time.Date(2026, 10, 2, 3, 0, 0, 0, time.UTC)
	order := Order{Status: StatusReady, PickupAt: &pickup}

	assert.False(t, order.ReadyLate(pickup), "ready at the pickup time is in time")
	assert.True(t, order.ReadyLate(pickup.Add(time.Minute)))
	assert.False(t, Order{Status: StatusInProduction, PickupAt: &pickup}.ReadyLate(pickup.Add(time.Hour)))
	assert.False(t, Order{Status: StatusReady}.ReadyLate(pickup), "no pickup time, never late")
}
