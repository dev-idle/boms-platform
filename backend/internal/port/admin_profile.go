package port

import (
	"context"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	"github.com/google/uuid"
)

type UpsertAdminProfileParams struct {
	UserID   uuid.UUID
	FullName string
	Phone    *string
}

type AdminProfileRepository interface {
	Create(ctx context.Context, params UpsertAdminProfileParams) (*domainprofile.Admin, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Admin, error)
	UpdateByUserID(ctx context.Context, params UpsertAdminProfileParams) (*domainprofile.Admin, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
