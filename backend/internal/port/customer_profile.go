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
	// MarketingOptIn gives (true) or withdraws (false) the customer's agreement
	// to promotions; nil leaves it as it is, and a new profile without it.
	MarketingOptIn *bool
}

type CustomerProfileRepository interface {
	// Erase clears the name and phone a customer gave; the row stays, empty.
	Erase(ctx context.Context, userID uuid.UUID) error
	Create(ctx context.Context, params UpsertCustomerProfileParams) (*domainprofile.Customer, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Customer, error)
	UpdateByUserID(ctx context.Context, params UpsertCustomerProfileParams) (*domainprofile.Customer, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	// WithdrawMarketingConsent reports whether the customer had agreed to promotions.
	WithdrawMarketingConsent(ctx context.Context, userID uuid.UUID) (bool, error)
}
