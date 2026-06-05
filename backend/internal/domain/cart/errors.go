package cart

import "errors"

var (
	ErrItemNotFound        = errors.New("cart item not found")
	ErrEmpty               = errors.New("cart is empty")
	ErrProductUnavailable  = errors.New("product unavailable")
	ErrComboUnavailable    = errors.New("combo unavailable")
	ErrInvalidLine         = errors.New("invalid cart line")
	ErrMaxItemsReached     = errors.New("cart max items reached")
	ErrQuantityOutOfRange  = errors.New("cart quantity out of range")
)
