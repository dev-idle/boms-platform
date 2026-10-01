package product

import "errors"

var (
	ErrNotFound   = errors.New("product not found")
	ErrSlugExists = errors.New("product slug exists")
	// ErrInvalidOption refuses an option a manager offers: a group, a label of
	// 1 to 60 characters of plain text, and an added price of 0 to $1,000.
	ErrInvalidOption = errors.New("invalid product option")
	// ErrInvalidChoice refuses a customer's choice of options: one active
	// option from each group the product offers.
	ErrInvalidChoice = errors.New("invalid option choice")
	// ErrInvalidMessage refuses a cake message over 60 characters or not plain text.
	ErrInvalidMessage = errors.New("invalid cake message")
	// ErrCustomization refuses a cart line configured for a product that is not
	// customizable, or not configured for one that is.
	ErrCustomization = errors.New("customization does not match the product")
)
