package discount

import "errors"

var (
	ErrNotFound        = errors.New("discount code not found")
	ErrCodeExists      = errors.New("discount code exists")
	ErrInvalidCode     = errors.New("discount code invalid")
	ErrInactive        = errors.New("discount code inactive")
	ErrExpired         = errors.New("discount code expired")
	ErrExhausted       = errors.New("discount code exhausted")
	ErrMinOrderNotMet  = errors.New("discount min order not met")
)
