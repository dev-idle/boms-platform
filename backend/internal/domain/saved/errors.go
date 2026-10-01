package saved

import "errors"

var (
	ErrInvalidList = errors.New("saved: no such list")
	ErrListFull    = errors.New("saved: the list is full")
)
