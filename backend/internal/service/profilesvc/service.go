package profilesvc

import (
	"context"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/google/uuid"
)

type Service struct {
	customers port.CustomerProfileRepository
	staff     port.StaffProfileRepository
	admins    port.AdminProfileRepository
}

func NewService(
	customers port.CustomerProfileRepository,
	staff port.StaffProfileRepository,
	admins port.AdminProfileRepository,
) *Service {
	return &Service{
		customers: customers,
		staff:     staff,
		admins:    admins,
	}
}

// GetByUserID returns a role-matched profile entity.
func (s *Service) GetByUserID(ctx context.Context, userID uuid.UUID, role domainuser.Role) (any, error) {
	switch role.ProfileType() {
	case "customer":
		return s.customers.GetByUserID(ctx, userID)
	case "staff":
		return s.staff.GetByUserID(ctx, userID)
	case "admin":
		return s.admins.GetByUserID(ctx, userID)
	default:
		return nil, domainuser.ErrProfileNotFound
	}
}
