// Package user defines the user aggregate for authentication and authorization.
package user

import (
	"time"

	"github.com/google/uuid"
)

// User is the domain user entity (persistence-agnostic).
type User struct {
	ID                 uuid.UUID
	Email              string
	PasswordHash       string
	Role               Role
	EmailVerified      bool
	MustChangePassword bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
	// ErasedAt is set once the account's owner had its personal details erased;
	// such an account is closed for good.
	ErasedAt *time.Time
	// SessionVersion goes up whenever every session of the account must end (a
	// new password, disabling it, an administrator's revoke); a session carries the version
	// it began under and cannot refresh past a newer one. Filled only by
	// GetByEmail and GetByID; other lookups leave it 0.
	SessionVersion int32
}

// Disabled reports whether the user was soft-deleted.
func (u User) Disabled() bool {
	return u.DeletedAt != nil
}

// Erased reports whether the account was erased at its owner's request.
func (u User) Erased() bool {
	return u.ErasedAt != nil
}
