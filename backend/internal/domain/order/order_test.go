package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
