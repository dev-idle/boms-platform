package port

import (
	"context"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	"github.com/google/uuid"
)

type UpsertCustomerProfileParams struct {
	UserID      uuid.UUID
	DisplayName *string
	Phone       *string
}

type CustomerProfileRepository interface {
	Create(ctx context.Context, params UpsertCustomerProfileParams) (*domainprofile.Customer, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Customer, error)
	UpdateByUserID(ctx context.Context, params UpsertCustomerProfileParams) (*domainprofile.Customer, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
