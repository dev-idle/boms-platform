package conversation

import "errors"

var (
	ErrInvalidBody = errors.New("conversation: a message is 1 to 2000 plain characters")
	// ErrUnavailable: the order cannot be written about — the bakery has not
	// taken it yet, or it was taken for a guest, who has no account to answer from.
	ErrUnavailable = errors.New("conversation: messages are not available for this order")
	ErrNotFound    = errors.New("conversation: not found")
)
