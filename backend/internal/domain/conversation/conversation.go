// Package conversation is a customer and the counter writing to each other
// about one of the customer's orders.
package conversation

import (
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
)

// MaxBodyLength is the most characters one message holds.
const MaxBodyLength = 2000

// Status is whether a conversation still waits on the counter.
type Status string

const (
	StatusOpen   Status = "open"
	StatusClosed Status = "closed"
)

// Conversation is where an order's conversation stands. AssignedStaffName is
// the staff member who replied last, nil until the counter answers.
type Conversation struct {
	Status            Status
	UnreadByCustomer  int32
	UnreadByStaff     int32
	AssignedStaffName *string
}

// Message is one message. AuthorRole is the customer or staff; AuthorName is
// a staff writer's name, nil for the customer.
type Message struct {
	ID         uuid.UUID
	AuthorRole domainuser.Role
	AuthorName *string
	Body       string
	CreatedAt  time.Time
}

// Valid reports whether s is a status this package defines.
func (s Status) Valid() bool {
	return s == StatusOpen || s == StatusClosed
}

// NewBody checks a message as written: trimmed, 1 to MaxBodyLength
// characters, plain text that may break across lines.
func NewBody(body string) (string, error) {
	body = domainstore.TrimLines(body)
	if body == "" || utf8.RuneCountInString(body) > MaxBodyLength || !domainstore.PlainLines(body) {
		return "", ErrInvalidBody
	}
	return body, nil
}
