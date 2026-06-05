package product

import "errors"

var (
	ErrNotFound   = errors.New("product not found")
	ErrSlugExists = errors.New("product slug exists")
)
