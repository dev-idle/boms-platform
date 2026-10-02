package review

import "errors"

var (
	ErrNotFound       = errors.New("review: not found")
	ErrInvalidRating  = errors.New("review: a rating is 1 to 5 stars")
	ErrInvalidComment = errors.New("review: a comment is up to 1000 plain characters")
	ErrInvalidStatus  = errors.New("review: a manager publishes or hides a review")
	// ErrNotReviewable: the product was not on the order, or the order was
	// not picked up.
	ErrNotReviewable = errors.New("review: only a product on a picked-up order is reviewed")
	ErrExists        = errors.New("review: the product on this order is already reviewed")
)
