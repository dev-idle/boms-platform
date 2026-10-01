package port

import (
	"context"
	"time"

	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/google/uuid"
)

// CreateUserParams holds fields required to register a user.
type CreateUserParams struct {
	Email              string
	PasswordHash       string
	Role               domainuser.Role
	MustChangePassword bool
	// TermsVersion is the policy version accepted at sign-up; accounts an admin
	// creates have none.
	TermsVersion *string
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
	ErasedAt           *time.Time
	FullName           *string
	Phone              *string
	EmployeeCode       *string
	DisplayName        *string
}

// CustomerContact is a customer account staff take an order for.
type CustomerContact struct {
	ID          uuid.UUID
	Email       string
	DisplayName *string
	Phone       *string
}

// UserRepository loads and persists users without auth policy.
type UserRepository interface {
	// FindCustomerByEmail is the open customer account with that email; any
	// other is apperrors.ErrNotFound.
	FindCustomerByEmail(ctx context.Context, email string) (*CustomerContact, error)
	Create(ctx context.Context, params CreateUserParams) (*domainuser.User, error)
	AdminCreate(ctx context.Context, params CreateUserParams) (*domainuser.User, error)
	GetByEmail(ctx context.Context, email string) (*domainuser.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	// GetByIDForShare is GetByID with the account held open until the
	// transaction ends: closing it waits, and a read that waited on a closing
	// finds the account gone.
	GetByIDForShare(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateRole(ctx context.Context, id uuid.UUID, role domainuser.Role) error
	SetMustChangePassword(ctx context.Context, id uuid.UUID) error
	ClearMustChangePassword(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	AdminGetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	// AdminGetByIDForUpdate is AdminGetByID with the row locked until the transaction ends.
	AdminGetByIDForUpdate(ctx context.Context, id uuid.UUID) (*domainuser.User, error)
	Restore(ctx context.Context, id uuid.UUID) error
	AdminUpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	AdminList(ctx context.Context, params AdminListUsersParams) ([]AdminListUser, error)
	AdminListCount(ctx context.Context, search string, role *domainuser.Role) (int64, error)
	// ClaimPhone locks phone until the surrounding transaction ends and reports
	// whether an active account other than userID already holds it.
	ClaimPhone(ctx context.Context, phone string, userID uuid.UUID) (held bool, err error)
	// ReleasePhone clears the phone on the user's profile, whichever type it is.
	ReleasePhone(ctx context.Context, userID uuid.UUID) error
	// Erase clears the account's personal details and closes it for good; the
	// row stays for the orders it placed. It refuses an account already closed.
	Erase(ctx context.Context, id uuid.UUID) error
	// MarkEmailVerified records the account's address as confirmed; confirming
	// it again keeps the first time.
	MarkEmailVerified(ctx context.Context, id uuid.UUID) error
	// ResetPassword sets a new password from an emailed reset link: it lifts a
	// required change and, as the link reached the inbox, confirms the address.
	ResetPassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	// BumpSessionVersion ends every session of the account for good: none can
	// refresh again. Every password write and SoftDelete bump it too.
	BumpSessionVersion(ctx context.Context, id uuid.UUID) error
	// TermsAcceptance is the policy version the user accepted at sign-up, or nil.
	TermsAcceptance(ctx context.Context, userID uuid.UUID) (*domainpolicy.Acceptance, error)
}
