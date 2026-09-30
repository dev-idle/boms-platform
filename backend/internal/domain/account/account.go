// Package account holds the rules of the single-use links a user gets by
// email: confirming their address and choosing a new password.
package account

import (
	"errors"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
)

// ErrLinkInvalid refuses a link that is unknown, expired or already used, or
// whose account is closed. The four read the same, so a link tells nobody
// anything about the account behind it.
var ErrLinkInvalid = errors.New("link invalid")

// Purpose is what a link does.
type Purpose string

const (
	PurposeVerifyEmail   Purpose = "verify_email"
	PurposeResetPassword Purpose = "reset_password"
)

// Valid reports whether p is a purpose this package defines.
func (p Purpose) Valid() bool {
	return p == PurposeVerifyEmail || p == PurposeResetPassword
}

// TTL is how long a link works from when it is sent. A confirmation waits for
// someone who may read their mail tomorrow; a password reset is used at once,
// and a short life limits what a leaked link is worth.
func (p Purpose) TTL() time.Duration {
	if p == PurposeResetPassword {
		return 30 * time.Minute
	}
	return 48 * time.Hour
}

const (
	// TopicVerificationRequested asks for an email confirming a user's address.
	TopicVerificationRequested domainevent.Topic = "account.verification_requested"
	// TopicPasswordResetRequested asks for an email with a password reset link.
	TopicPasswordResetRequested domainevent.Topic = "account.password_reset_requested"
)

// VerificationRequestedEvent asks for userID's address to be confirmed. Its
// audience is nobody: it reaches the email queue, never an open page.
func VerificationRequestedEvent(userID uuid.UUID) domainevent.Event {
	return requestEvent(TopicVerificationRequested, userID)
}

// PasswordResetRequestedEvent asks for a reset link for userID. The link is
// made when the email is sent, so neither the event nor the queue holds one.
func PasswordResetRequestedEvent(userID uuid.UUID) domainevent.Event {
	return requestEvent(TopicPasswordResetRequested, userID)
}

func requestEvent(topic domainevent.Topic, userID uuid.UUID) domainevent.Event {
	return domainevent.New(topic, domainevent.Audience{}, map[string]string{"user_id": userID.String()})
}

// EmailFor names the account email an event calls for, and its user.
func EmailFor(e domainevent.Event) (uuid.UUID, Purpose, bool) {
	var purpose Purpose
	switch e.Topic {
	case TopicVerificationRequested:
		purpose = PurposeVerifyEmail
	case TopicPasswordResetRequested:
		purpose = PurposeResetPassword
	default:
		return uuid.Nil, "", false
	}
	userID, err := uuid.Parse(e.Data["user_id"])
	if err != nil {
		return uuid.Nil, "", false
	}
	return userID, purpose, true
}
