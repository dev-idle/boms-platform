package combo

import "errors"

var (
	ErrNotFound   = errors.New("combo not found")
	ErrSlugExists = errors.New("combo slug exists")
)
