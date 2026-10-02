package profile

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	UserID      uuid.UUID
	DisplayName *string
	Phone       *string
	// MarketingConsentAt is when the customer agreed to be emailed
	// promotions; nil while they have not, or since they withdrew.
	MarketingConsentAt *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
