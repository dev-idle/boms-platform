package profile

import (
	"time"

	"github.com/google/uuid"
)

type Admin struct {
	UserID    uuid.UUID
	FullName  string
	Phone     *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
