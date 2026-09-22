package usecase_test

import (
	"context"
	"errors"
	"testing"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const storedPhone = "+84912345678"

func phonePtr(value string) *string {
	return &value
}

func TestPhoneClaim(t *testing.T) {
	t.Parallel()

	t.Run("create_operational_rejects_a_held_phone", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		hasher := new(mockHasher)
		uc := usecase.NewAdminUserUsecase(users, nil, nil, nil, nil, passthroughTxManager{}, hasher, nil, nil, nil)
		created := &domainuser.User{ID: uuid.New(), Role: domainuser.RoleStaff}

		hasher.On("Hash", mock.Anything).Return("temp-hash", nil).Once()
		users.On("AdminCreate", mock.Anything, mock.Anything).Return(created, nil).Once()
		users.On("ClaimPhone", mock.Anything, storedPhone, created.ID).Return(true, nil).Once()

		_, err := uc.CreateOperationalUser(context.Background(), uuid.New(), domainuser.RoleAdmin, dto.CreateOperationalUserRequest{
			Email:    "new@example.com",
			Role:     string(domainuser.RoleStaff),
			FullName: "New Staff",
			Phone:    phonePtr("0912 345 678"),
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, domainuser.ErrPhoneExists), "got %v", err)
		users.AssertExpectations(t)
	})

	t.Run("self_update_claims_the_normalized_phone", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		customers := new(mockCustomerProfileRepo)
		uc := usecase.NewMeUsecase(users, customers, nil, nil, nil, passthroughTxManager{}, nil, nil, nil)
		userID := uuid.New()

		users.On("GetByID", mock.Anything, userID).Return(&domainuser.User{ID: userID, Role: domainuser.RoleCustomer}, nil).Once()
		customers.On("GetByUserID", mock.Anything, userID).Return(&domainprofile.Customer{UserID: userID}, nil).Once()
		users.On("ClaimPhone", mock.Anything, storedPhone, userID).Return(true, nil).Once()

		_, _, err := uc.UpdateProfile(context.Background(), userID, dto.UpdateMeRequest{Phone: phonePtr("+84 912-345-678")})

		require.Error(t, err)
		assert.True(t, errors.Is(err, domainuser.ErrPhoneExists), "got %v", err)
		users.AssertExpectations(t)
		customers.AssertNotCalled(t, "UpdateByUserID", mock.Anything, mock.Anything)
	})

	t.Run("self_update_does_not_reclaim_an_unchanged_phone", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		customers := new(mockCustomerProfileRepo)
		uc := usecase.NewMeUsecase(users, customers, nil, nil, nil, passthroughTxManager{}, nil, nil, nil)
		userID := uuid.New()
		current := &domainprofile.Customer{UserID: userID, Phone: phonePtr(storedPhone)}

		users.On("GetByID", mock.Anything, userID).Return(&domainuser.User{ID: userID, Role: domainuser.RoleCustomer}, nil).Once()
		customers.On("GetByUserID", mock.Anything, userID).Return(current, nil).Once()
		customers.On("UpdateByUserID", mock.Anything, port.UpsertCustomerProfileParams{
			UserID:      userID,
			DisplayName: phonePtr("Pat"),
			Phone:       phonePtr(storedPhone),
		}).Return(current, nil).Once()

		_, _, err := uc.UpdateProfile(context.Background(), userID, dto.UpdateMeRequest{
			DisplayName: phonePtr("Pat"),
			Phone:       phonePtr("0912 345 678"),
		})

		require.NoError(t, err)
		users.AssertNotCalled(t, "ClaimPhone", mock.Anything, mock.Anything, mock.Anything)
		customers.AssertExpectations(t)
	})

	t.Run("self_update_fails_on_a_phone_the_validator_would_reject", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		uc := usecase.NewMeUsecase(users, nil, nil, nil, nil, passthroughTxManager{}, nil, nil, nil)

		_, _, err := uc.UpdateProfile(context.Background(), uuid.New(), dto.UpdateMeRequest{Phone: phonePtr("!!!!!!")})

		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr), "got %v", err)
		assert.Equal(t, "vn_phone", appErr.Details["phone"])
		users.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	})

	t.Run("enable_releases_a_phone_taken_while_disabled", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		staff := new(mockStaffProfileRepo)
		uc := usecase.NewAdminUserUsecase(users, nil, staff, nil, nil, passthroughTxManager{}, nil, nil, nil, nil)
		targetID := uuid.New()

		users.On("AdminGetByID", mock.Anything, targetID).Return(&domainuser.User{ID: targetID, Role: domainuser.RoleStaff}, nil).Once()
		users.On("Restore", mock.Anything, targetID).Return(nil).Once()
		staff.On("GetByUserID", mock.Anything, targetID).Return(&domainprofile.Staff{UserID: targetID, Phone: phonePtr(storedPhone)}, nil).Once()
		users.On("ClaimPhone", mock.Anything, storedPhone, targetID).Return(true, nil).Once()
		users.On("ReleasePhone", mock.Anything, targetID).Return(nil).Once()

		err := uc.Enable(context.Background(), uuid.New(), domainuser.RoleAdmin, targetID)

		require.NoError(t, err, "the account comes back; the active holder keeps the number")
		users.AssertExpectations(t)
	})

	t.Run("enable_keeps_a_phone_nobody_took", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		staff := new(mockStaffProfileRepo)
		uc := usecase.NewAdminUserUsecase(users, nil, staff, nil, nil, passthroughTxManager{}, nil, nil, nil, nil)
		targetID := uuid.New()

		users.On("AdminGetByID", mock.Anything, targetID).Return(&domainuser.User{ID: targetID, Role: domainuser.RoleStaff}, nil).Once()
		users.On("Restore", mock.Anything, targetID).Return(nil).Once()
		staff.On("GetByUserID", mock.Anything, targetID).Return(&domainprofile.Staff{UserID: targetID, Phone: phonePtr(storedPhone)}, nil).Once()
		users.On("ClaimPhone", mock.Anything, storedPhone, targetID).Return(false, nil).Once()

		require.NoError(t, uc.Enable(context.Background(), uuid.New(), domainuser.RoleAdmin, targetID))
		users.AssertNotCalled(t, "ReleasePhone", mock.Anything, mock.Anything)
	})
}
