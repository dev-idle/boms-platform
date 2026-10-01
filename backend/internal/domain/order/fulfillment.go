package order

import "time"

// Type is how an order is prepared. Instant orders are ready-made counter items
// collected the day they are placed; every other order is a pre-order.
type Type string

const (
	TypeInstant  Type = "instant"
	TypePreOrder Type = "pre_order"
)

// Fulfillment is what an order's items ask of the bakery: whether any is made
// in the kitchen, the longest notice any of them needs, and the bakery day one
// of them has run out on (today, or nil), when the order cannot be collected.
type Fulfillment struct {
	Kitchen   bool
	Lead      time.Duration
	SoldOutOn *time.Time
}

// HeldOn is what the items of an order collected on day ask of the bakery that
// day: the order holds them already, so an item sold out since does not stop it.
func (f Fulfillment) HeldOn(day time.Time) Fulfillment {
	if f.SoldOutOn != nil && f.SoldOutOn.Equal(day) {
		f.SoldOutOn = nil
	}
	return f
}
