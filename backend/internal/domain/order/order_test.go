package order

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrder_Payable(t *testing.T) {
	t.Parallel()
	now := time.Now()
	due := now.Add(time.Minute)
	awaiting := Order{Status: StatusAwaitingPayment, PaymentDueAt: &due}

	assert.True(t, awaiting.Payable(now))
	assert.False(t, awaiting.Payable(due), "the time to pay is up")
	assert.False(t, Order{Status: StatusExpired, PaymentDueAt: &due}.Payable(now))
	assert.False(t, Order{Status: StatusPending}.Payable(now), "placed before online payment")
}

func TestOrder_HasPickupCode(t *testing.T) {
	t.Parallel()
	assert.True(t, Order{Channel: ChannelOnline, Status: StatusReady}.HasPickupCode())
	assert.False(t, Order{Channel: ChannelCounter, Status: StatusReady}.HasPickupCode(), "handed over as its cash is paid")
	assert.False(t, Order{Channel: ChannelPhone, Status: StatusConfirmed}.HasPickupCode())
	assert.False(t, Order{Channel: ChannelOnline, Status: StatusAwaitingPayment}.HasPickupCode())
}

// An order holds its items on its own day: an item sold out since does not
// stop it there, and does stop any other day.
func TestFulfillment_HeldOn(t *testing.T) {
	t.Parallel()
	today := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	items := Fulfillment{Kitchen: true, SoldOutOn: &today}

	assert.Nil(t, items.HeldOn(today).SoldOutOn)
	assert.Equal(t, &today, items.HeldOn(today.AddDate(0, 0, 1)).SoldOutOn)
	assert.True(t, items.HeldOn(today).Kitchen)
}

func TestNewGuest(t *testing.T) {
	t.Parallel()
	guest, err := NewGuest("  Lan Nguyen ", "+84901234567")
	require.NoError(t, err)
	assert.Equal(t, Guest{Name: "Lan Nguyen", Phone: "+84901234567"}, guest)

	for _, name := range []string{"", "   ", strings.Repeat("a", maxGuestNameLength+1), "Lan" + string(rune(0x202e)) + "Nguyen"} {
		_, err := NewGuest(name, "+84901234567")
		assert.ErrorIs(t, err, ErrInvalidGuestName, "%q", name)
	}
}
