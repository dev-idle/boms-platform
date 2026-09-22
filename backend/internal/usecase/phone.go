package usecase

import (
	"context"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
)

// normalizeRequestPhone puts a request phone in its stored form, keeping PATCH
// meaning: nil is absent, "" clears. The handler's vn_phone tag has already run;
// failing here instead of storing the raw value keeps the usecase zero-trust.
func normalizeRequestPhone(phone *string) (*string, error) {
	normalized, ok := utils.NormalizeVietnamPhonePtr(phone)
	if !ok {
		return nil, apperrors.ErrValidation.WithDetail("phone", "vn_phone")
	}
	return normalized, nil
}

// claimPhone refuses a phone that another active account already holds. It must
// run inside the transaction that writes the phone: the lock it takes is what
// stops two concurrent writers from both finding the number free.
//
// A number the account already holds is not claimed again. Two accounts may share
// one from before the rule existed; re-checking an unchanged phone would lock
// both out of saving anything else on their profile.
func claimPhone(ctx context.Context, users port.UserRepository, userID uuid.UUID, next, current *string) error {
	if next == nil || *next == "" {
		return nil
	}
	if current != nil && *current == *next {
		return nil
	}
	held, err := users.ClaimPhone(ctx, *next, userID)
	if err != nil {
		return err
	}
	if held {
		return domainuser.ErrPhoneExists
	}
	return nil
}

// profilePhone reads the phone of any profile type; nil when there is none.
func profilePhone(profile any) *string {
	switch p := profile.(type) {
	case *domainprofile.Customer:
		return p.Phone
	case *domainprofile.Staff:
		return p.Phone
	case *domainprofile.Admin:
		return p.Phone
	default:
		return nil
	}
}
