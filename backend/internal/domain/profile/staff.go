package profile

import (
	"time"

	"github.com/google/uuid"
)

type Staff struct {
	UserID       uuid.UUID
	FullName     string
	Phone        *string
	EmployeeCode string
	HireDate     time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
