package port

import (
	"context"
	"time"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/google/uuid"
)

// CreateUserParams holds fields required to register a user.
type CreateUserParams struct {
	Email              string
	PasswordHash       string
	Role               domainuser.Role
	MustChangePassword bool
}

type AdminListUsersParams struct {
	Search string
	Role   *domainuser.Role
	Limit  int32
	Offset int32
}

type AdminListUser struct {
	ID                 uuid.UUID
	Email              string
	Role               domainuser.Role
	EmailVerified      bool
	MustChangePassword bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
	FullName           *string
	Phone              *string
	EmployeeCode       *string
	DisplayName        *string
}

// UserRepository loads and persists users without auth policy.
type UserRepository interface {
	Create(ctx context.Context, params CreateUserParams) (*domainuser.User, error)
	AdminCreate(ctx context.Context, params CreateUserParams) (*domainuser.User, error)
	GetByEmail(ctx context.Context, email string) (*domainuser.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateRole(ctx context.Context, id uuid.UUID, role domainuser.Role) error
	SetMustChangePassword(ctx context.Context, id uuid.UUID) error
	ClearMustChangePassword(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	AdminGetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	Restore(ctx context.Context, id uuid.UUID) error
	AdminUpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	AdminList(ctx context.Context, params AdminListUsersParams) ([]AdminListUser, int64, error)
}
