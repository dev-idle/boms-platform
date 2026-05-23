package port

import (
	"context"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/google/uuid"
)

// CreateUserParams holds fields required to register a user.
type CreateUserParams struct {
	Email        string
	PasswordHash string
	Role         domainuser.Role
}

// UserRepository loads and persists users without auth policy.
type UserRepository interface {
	Create(ctx context.Context, params CreateUserParams) (*domainuser.User, error)
	GetByEmail(ctx context.Context, email string) (*domainuser.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
