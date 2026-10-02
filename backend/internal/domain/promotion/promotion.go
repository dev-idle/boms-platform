// Package promotion is a manager emailing a promotion to the customers who
// agreed to receive promotions, and a customer withdrawing that agreement from
// any promotion email.
package promotion

import (
	"strings"
	"unicode/utf8"

	domainstore "github.com/boms/backend/internal/domain/store"
)

const (
	MaxSubjectLength = 120
	MaxBodyLength    = 5000
)

// Status is where a promotion stands: sending while the worker queues an email
// for each customer it goes to, then sent.
type Status string

const (
	StatusSending Status = "sending"
	StatusSent    Status = "sent"
)

// NewSubject checks a promotion's subject: trimmed, 1 to MaxSubjectLength
// plain characters on one line, as an email subject is.
func NewSubject(subject string) (string, error) {
	subject = strings.TrimSpace(subject)
	if subject == "" || utf8.RuneCountInString(subject) > MaxSubjectLength || !domainstore.PlainText(subject) {
		return "", ErrInvalidSubject
	}
	return subject, nil
}

// NewBody checks a promotion's message: trimmed, 1 to MaxBodyLength plain
// characters that may break across lines.
func NewBody(body string) (string, error) {
	body = domainstore.TrimLines(body)
	if body == "" || utf8.RuneCountInString(body) > MaxBodyLength || !domainstore.PlainLines(body) {
		return "", ErrInvalidBody
	}
	return body, nil
}
