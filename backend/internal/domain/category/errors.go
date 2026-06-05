package category

import "errors"

var (
	ErrNotFound          = errors.New("category not found")
	ErrHasProducts       = errors.New("category has products")
	ErrSlugExists        = errors.New("category slug exists")
	ErrInactive          = errors.New("category inactive")
)
