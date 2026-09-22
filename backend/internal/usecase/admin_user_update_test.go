package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestAdminUserUsecase_RefusesAdminTargets(t *testing.T) {
	t.Parallel()

	newAdminTarget := func() (*usecase.AdminUserUsecase, *mockUserRepo, uuid.UUID) {
		users := new(mockUserRepo)
		targetID := uuid.New()
		admin := &domainuser.User{ID: targetID, Role: domainuser.RoleAdmin}
		users.On("AdminGetByID", mock.Anything, targetID).Return(admin, nil).Maybe()
		users.On("AdminGetByIDForUpdate", mock.Anything, targetID).Return(admin, nil).Maybe()
		return usecase.NewAdminUserUsecase(users, nil, nil, nil, nil, passthroughTxManager{}, nil, nil, nil, nil), users, targetID
	}

	t.Run("profile_update", func(t *testing.T) {
		t.Parallel()
		uc, _, targetID := newAdminTarget()

		_, err := uc.UpdateOperationalProfile(context.Background(), uuid.New(), domainuser.RoleAdmin, targetID, dto.UpdateOperationalProfileRequest{FullName: "Other Admin"})

		assert.True(t, errors.Is(err, domainuser.ErrCannotModifyAdmin), "got %v", err)
	})

	t.Run("role_change", func(t *testing.T) {
		t.Parallel()
		uc, users, targetID := newAdminTarget()

		_, err := uc.UpdateRole(context.Background(), uuid.New(), domainuser.RoleAdmin, targetID, dto.UpdateUserRoleRequest{Role: string(domainuser.RoleStaff)})

		assert.True(t, errors.Is(err, domainuser.ErrCannotModifyAdmin), "got %v", err)
		users.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("password_reset", func(t *testing.T) {
		t.Parallel()
		uc, users, targetID := newAdminTarget()

		out, err := uc.ResetPassword(context.Background(), uuid.New(), domainuser.RoleAdmin, targetID)

		assert.Nil(t, out, "no temporary password may leave for another admin")
		assert.True(t, errors.Is(err, domainuser.ErrCannotModifyAdmin), "got %v", err)
		users.AssertNotCalled(t, "AdminUpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestAdminUserUsecase_UpdateRole(t *testing.T) {
	t.Parallel()

	type fixture struct {
		uc       *usecase.AdminUserUsecase
		users    *mockUserRepo
		staff    *mockStaffProfileRepo
		sessions *mockSessionStore
		targetID uuid.UUID
		current  *domainprofile.Staff
	}
	newFixture := func(role domainuser.Role) fixture {
		users := new(mockUserRepo)
		staff := new(mockStaffProfileRepo)
		sessions := new(mockSessionStore)
		targetID := uuid.New()
		current := &domainprofile.Staff{UserID: targetID, FullName: "Linh Tran", EmployeeCode: "EMP-00012"}
		target := &domainuser.User{ID: targetID, Role: role}
		users.On("AdminGetByIDForUpdate", mock.Anything, targetID).Return(target, nil).Maybe()
		users.On("AdminGetByID", mock.Anything, targetID).Return(target, nil).Maybe()
		staff.On("GetByUserID", mock.Anything, targetID).Return(current, nil).Maybe()
		return fixture{
			uc:       usecase.NewAdminUserUsecase(users, nil, staff, nil, sessions, passthroughTxManager{}, nil, nil, nil, nil),
			users:    users,
			staff:    staff,
			sessions: sessions,
			targetID: targetID,
			current:  current,
		}
	}

	t.Run("same_role_saves_name_and_phone_in_place", func(t *testing.T) {
		t.Parallel()
		f := newFixture(domainuser.RoleStaff)
		f.users.On("ClaimPhone", mock.Anything, storedPhone, f.targetID).Return(false, nil).Once()
		f.staff.On("UpdateByUserID", mock.Anything, port.UpsertStaffProfileParams{
			UserID:       f.targetID,
			FullName:     "Linh Tran Thi",
			Phone:        phonePtr(storedPhone),
			EmployeeCode: "EMP-00012",
		}).Return(f.current, nil).Once()

		_, err := f.uc.UpdateRole(context.Background(), uuid.New(), domainuser.RoleAdmin, f.targetID, dto.UpdateUserRoleRequest{
			Role:     string(domainuser.RoleStaff),
			FullName: "Linh Tran Thi",
			Phone:    phonePtr("0912 345 678"),
		})

		require.NoError(t, err)
		f.staff.AssertExpectations(t)
		f.users.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
		f.sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
	})

	t.Run("omitted_phone_keeps_the_stored_one", func(t *testing.T) {
		t.Parallel()
		f := newFixture(domainuser.RoleStaff)
		f.current.Phone = phonePtr(storedPhone)
		f.staff.On("UpdateByUserID", mock.Anything, port.UpsertStaffProfileParams{
			UserID:       f.targetID,
			FullName:     "Linh Tran",
			Phone:        phonePtr(storedPhone),
			EmployeeCode: "EMP-00012",
		}).Return(f.current, nil).Once()

		_, err := f.uc.UpdateRole(context.Background(), uuid.New(), domainuser.RoleAdmin, f.targetID, dto.UpdateUserRoleRequest{
			Role: string(domainuser.RoleStaff),
		})

		require.NoError(t, err)
		f.staff.AssertExpectations(t)
		f.users.AssertNotCalled(t, "ClaimPhone", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("role_change_keeps_the_employee_code_and_revokes_sessions", func(t *testing.T) {
		t.Parallel()
		f := newFixture(domainuser.RoleStaff)
		f.staff.On("UpdateByUserID", mock.Anything, mock.MatchedBy(func(p port.UpsertStaffProfileParams) bool {
			return p.EmployeeCode == "EMP-00012"
		})).Return(f.current, nil).Once()
		f.users.On("UpdateRole", mock.Anything, f.targetID, domainuser.RoleBaker).Return(nil).Once()
		f.sessions.On("DeleteAllForUser", mock.Anything, f.targetID.String()).Return(nil).Once()

		_, err := f.uc.UpdateRole(context.Background(), uuid.New(), domainuser.RoleAdmin, f.targetID, dto.UpdateUserRoleRequest{
			Role: string(domainuser.RoleBaker),
		})

		require.NoError(t, err)
		f.users.AssertExpectations(t)
		f.sessions.AssertExpectations(t)
		f.staff.AssertNotCalled(t, "DeleteByUserID", mock.Anything, mock.Anything)
		f.staff.AssertNotCalled(t, "NextEmployeeCode", mock.Anything)
	})

	t.Run("held_phone_blocks_the_whole_change", func(t *testing.T) {
		t.Parallel()
		f := newFixture(domainuser.RoleStaff)
		f.users.On("ClaimPhone", mock.Anything, storedPhone, f.targetID).Return(true, nil).Once()

		_, err := f.uc.UpdateRole(context.Background(), uuid.New(), domainuser.RoleAdmin, f.targetID, dto.UpdateUserRoleRequest{
			Role:  string(domainuser.RoleManager),
			Phone: phonePtr("0912345678"),
		})

		assert.True(t, errors.Is(err, domainuser.ErrPhoneExists), "got %v", err)
		f.staff.AssertNotCalled(t, "UpdateByUserID", mock.Anything, mock.Anything)
		f.users.AssertNotCalled(t, "UpdateRole", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("disabled_target_is_refused", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		targetID := uuid.New()
		disabledAt := time.Now()
		users.On("AdminGetByIDForUpdate", mock.Anything, targetID).Return(&domainuser.User{ID: targetID, Role: domainuser.RoleStaff, DeletedAt: &disabledAt}, nil).Once()
		uc := usecase.NewAdminUserUsecase(users, nil, nil, nil, nil, passthroughTxManager{}, nil, nil, nil, nil)

		_, err := uc.UpdateRole(context.Background(), uuid.New(), domainuser.RoleAdmin, targetID, dto.UpdateUserRoleRequest{Role: string(domainuser.RoleBaker)})

		var appErr *apperrors.AppError
		require.True(t, errors.As(err, &appErr), "got %v", err)
		assert.Equal(t, "user is disabled", appErr.Details["account"])
	})

	t.Run("missing_staff_profile_is_reported_as_such", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		staff := new(mockStaffProfileRepo)
		targetID := uuid.New()
		users.On("AdminGetByIDForUpdate", mock.Anything, targetID).Return(&domainuser.User{ID: targetID, Role: domainuser.RoleStaff}, nil).Once()
		staff.On("GetByUserID", mock.Anything, targetID).Return(nil, apperrors.ErrNotFound).Once()
		uc := usecase.NewAdminUserUsecase(users, nil, staff, nil, nil, passthroughTxManager{}, nil, nil, nil, nil)

		_, err := uc.UpdateRole(context.Background(), uuid.New(), domainuser.RoleAdmin, targetID, dto.UpdateUserRoleRequest{Role: string(domainuser.RoleBaker)})

		assert.True(t, errors.Is(err, domainuser.ErrProfileNotFound), "got %v", err)
	})
}

func TestAdminUserUsecase_UpdateOperationalProfile_RejectsBlankName(t *testing.T) {
	t.Parallel()
	users := new(mockUserRepo)
	uc := usecase.NewAdminUserUsecase(users, nil, nil, nil, nil, passthroughTxManager{}, nil, nil, nil, nil)

	_, err := uc.UpdateOperationalProfile(context.Background(), uuid.New(), domainuser.RoleAdmin, uuid.New(), dto.UpdateOperationalProfileRequest{FullName: "   "})

	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "got %v", err)
	assert.Equal(t, "required", appErr.Details["full_name"])
	users.AssertNotCalled(t, "AdminGetByIDForUpdate", mock.Anything, mock.Anything)
}
