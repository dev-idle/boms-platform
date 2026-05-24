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
}

// Disabled reports whether the user was soft-deleted.
func (u User) Disabled() bool {
	return u.DeletedAt != nil
}
