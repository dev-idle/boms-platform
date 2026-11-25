package port

import (
	"context"
	"errors"
)

// ErrEmailUndeliverable marks an email no retry can send: the mail server
// refused the address itself (a mailbox or domain that does not exist), or it
// could not be written. Every other error from sending may pass on a retry.
var ErrEmailUndeliverable = errors.New("email undeliverable")

// Email is one message, ready to send.
type Email struct {
	To      string
	ToName  string
	Subject string
	Text    string
	HTML    string
	// ListUnsubscribe is the page that stops emails like this one, which
	// mail clients offer as their own unsubscribe button; empty for an email
	// the recipient asked for.
	ListUnsubscribe string
}

// Mailer sends email to one recipient.
type Mailer interface {
	Send(ctx context.Context, email Email) error
}
