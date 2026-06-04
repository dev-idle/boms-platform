package port

import (
	"context"
	"time"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	"github.com/google/uuid"
)

type UpsertStaffProfileParams struct {
	UserID       uuid.UUID
	FullName     string
	Phone        *string
	EmployeeCode string
	HireDate     time.Time
}

type StaffProfileRepository interface {
	Create(ctx context.Context, params UpsertStaffProfileParams) (*domainprofile.Staff, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Staff, error)
	UpdateByUserID(ctx context.Context, params UpsertStaffProfileParams) (*domainprofile.Staff, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
