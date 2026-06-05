package port

import (
	"context"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	"github.com/google/uuid"
)

type UpsertStaffProfileParams struct {
	UserID       uuid.UUID
	FullName     string
	Phone        *string
	EmployeeCode string
}

type StaffProfileRepository interface {
	Create(ctx context.Context, params UpsertStaffProfileParams) (*domainprofile.Staff, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Staff, error)
	UpdateByUserID(ctx context.Context, params UpsertStaffProfileParams) (*domainprofile.Staff, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
