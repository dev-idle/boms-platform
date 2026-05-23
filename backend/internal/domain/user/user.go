// Package user defines the user aggregate for authentication and authorization.
package user

import (
	"time"

	"github.com/google/uuid"
)

// Role identifies a user's application role.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleStaff    Role = "staff"
	RoleBaker    Role = "baker"
	RoleManager  Role = "manager"
	RoleAdmin    Role = "admin"
)

// User is the domain user entity (persistence-agnostic).
type User struct {
	ID            uuid.UUID
	Email         string
	PasswordHash  string
	Role          Role
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}
