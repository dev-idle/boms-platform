// Package review is a customer rating a product they picked up, and the
// manager deciding whether the storefront shows it: only a pickup is reviewed,
// and nothing is shown before a manager reads it, so a review cannot be faked.
package review

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domainstore "github.com/boms/backend/internal/domain/store"
)

const (
	MinRating = 1
	MaxRating = 5
	// MaxCommentLength is the most characters a review's comment holds.
	MaxCommentLength = 1000
)

// Status is where a review stands with the manager: pending until read, then
// published on the storefront or hidden from it.
type Status string

const (
	StatusPending   Status = "pending"
	StatusPublished Status = "published"
	StatusHidden    Status = "hidden"
)

// Review is a customer's rating of one product on one of their orders, with
// what they wrote, nil when they wrote nothing.
type Review struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Rating    int16
	Comment   *string
	Status    Status
	CreatedAt time.Time
}

// Valid reports whether s is a status this package defines.
func (s Status) Valid() bool {
	return s == StatusPending || s == StatusPublished || s == StatusHidden
}

// NewModeration checks the status a manager gives a review: published or
// hidden. A review read once is never pending again.
func NewModeration(status string) (Status, error) {
	s := Status(status)
	if s != StatusPublished && s != StatusHidden {
		return "", ErrInvalidStatus
	}
	return s, nil
}

// NewRating checks a rating: whole stars from MinRating to MaxRating.
func NewRating(rating int) (int16, error) {
	if rating < MinRating || rating > MaxRating {
		return 0, ErrInvalidRating
	}
	return int16(rating), nil
}

// NewComment checks what a customer wrote with their rating: trimmed, up to
// MaxCommentLength plain characters that may break across lines. Nothing
// written is nil.
func NewComment(comment *string) (*string, error) {
	if comment == nil {
		return nil, nil
	}
	trimmed := domainstore.TrimLines(*comment)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > MaxCommentLength || !domainstore.PlainLines(trimmed) {
		return nil, ErrInvalidComment
	}
	return &trimmed, nil
}
