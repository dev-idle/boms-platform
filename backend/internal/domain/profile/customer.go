package profile

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	UserID      uuid.UUID
	DisplayName *string
	Phone       *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
