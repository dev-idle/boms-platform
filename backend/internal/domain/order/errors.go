package order

import "errors"

var (
	ErrNotFound               = errors.New("order not found")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
)
