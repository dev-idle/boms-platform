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
// in the kitchen, and the longest notice any of them needs.
type Fulfillment struct {
	Kitchen bool
	Lead    time.Duration
}
