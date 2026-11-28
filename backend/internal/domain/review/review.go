// Package review is a customer rating a product they picked up: only a pickup
// is reviewed, so a review cannot be faked, and the storefront shows it from
// the moment it is written until a manager hides it.
package review

import (
	"strings"
	"time"
	"unicode"
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

// Status is whether the storefront shows a review: published from the moment
// it is written, hidden once a manager hides it.
type Status string

const (
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
	return s == StatusPublished || s == StatusHidden
}

// NewModeration checks the status a manager gives a review: hidden, or
// published again.
func NewModeration(status string) (Status, error) {
	s := Status(status)
	if !s.Valid() {
		return "", ErrInvalidStatus
	}
	return s, nil
}

// PublicName is how the storefront names a review's author: the first word of
// the name they gave and the initial of the last ("Mai Tran" reads "Mai T."),
// so a review has a person behind it without spelling out who. Nil when they
// gave no name.
func PublicName(displayName *string) *string {
	if displayName == nil {
		return nil
	}
	words := strings.Fields(*displayName)
	if len(words) == 0 {
		return nil
	}
	name := words[0]
	if len(words) > 1 {
		initial, _ := utf8.DecodeRuneInString(words[len(words)-1])
		name += " " + string(unicode.ToUpper(initial)) + "."
	}
	return &name
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
