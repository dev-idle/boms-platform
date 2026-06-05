package discount

import "errors"

var (
	ErrNotFound    = errors.New("discount code not found")
	ErrCodeExists  = errors.New("discount code exists")
	ErrInvalidCode = errors.New("discount code invalid")
)
