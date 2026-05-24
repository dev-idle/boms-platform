package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/service/profilesvc"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
)

type AdminUserUsecase struct {
	users     port.UserRepository
	customers port.CustomerProfileRepository
	staff     port.StaffProfileRepository
	admins    port.AdminProfileRepository
	sessions  port.SessionStore
	tx        port.TxManager
	hasher    port.PasswordHasher
	audit     *auditlogger.Service
	profiles  *profilesvc.Service
}

func NewAdminUserUsecase(
	users port.UserRepository,
	customers port.CustomerProfileRepository,
	staff port.StaffProfileRepository,
	admins port.AdminProfileRepository,
	sessions port.SessionStore,
	tx port.TxManager,
	hasher port.PasswordHasher,
	audit *auditlogger.Service,
) *AdminUserUsecase {
	return &AdminUserUsecase{
		users:     users,
		customers: customers,
		staff:     staff,
		admins:    admins,
		sessions:  sessions,
		tx:        tx,
		hasher:    hasher,
		audit:     audit,
		profiles:  profilesvc.NewService(customers, staff, admins),
	}
}

func (u *AdminUserUsecase) CreateOperationalUser(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.CreateOperationalUserRequest,
) (*dto.CreateOperationalUserResponse, error) {
	role, err := parseRole(req.Role)
	if err != nil {
		return nil, err
	}
	if !role.CanBeAssigned() || role.IsAdmin() {
		return nil, apperrors.ErrValidation.WithDetail("role", "must be staff, baker, or manager")
	}

	tempPassword, err := utils.GenerateTempPassword(16)
	if err != nil {
		return nil, err
	}
	hash, err := u.hasher.Hash(tempPassword)
	if err != nil {
		return nil, fmt.Errorf("hash temp password: %w", err)
	}

	var user *domainuser.User
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		created, createErr := u.users.AdminCreate(txCtx, port.CreateUserParams{
			Email:              strings.TrimSpace(strings.ToLower(req.Email)),
			PasswordHash:       hash,
			Role:               role,
			MustChangePassword: true,
		})
		if createErr != nil {
			return createErr
		}
		user = created

		switch role.ProfileType() {
		case "staff":
			employeeCode, hireDate, shift, prepErr := parseStaffSeed(req.EmployeeCode, req.HireDate, req.Shift)
			if prepErr != nil {
				return prepErr
			}
			_, createErr = u.staff.Create(txCtx, port.UpsertStaffProfileParams{
				UserID:       user.ID,
				FullName:     req.FullName,
				Phone:        req.Phone,
				EmployeeCode: employeeCode,
				HireDate:     hireDate,
				Shift:        shift,
			})
			return createErr
		case "admin":
			_, createErr = u.admins.Create(txCtx, port.UpsertAdminProfileParams{
				UserID:   user.ID,
				FullName: req.FullName,
				Phone:    req.Phone,
			})
			return createErr
		default:
			return domainuser.ErrInvalidRoleTransition
		}
	}); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domainuser.ErrEmployeeCodeExists
		}
		return nil, err
	}

	listRows, _, err := u.users.AdminList(ctx, port.AdminListUsersParams{
		Search: user.Email,
		Limit:  1,
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}
	if len(listRows) == 0 {
		return nil, apperrors.ErrNotFound
	}

	u.logAudit(ctx, domainuser.AuditActionAdminCreatedUser, actorID, actorRole, &user.ID, "user", nil, map[string]any{"role": req.Role})
	return &dto.CreateOperationalUserResponse{
		User:         mapAdminListItem(listRows[0]),
		TempPassword: tempPassword,
	}, nil
}

func (u *AdminUserUsecase) UpdateOperationalProfile(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	targetID uuid.UUID,
	req dto.UpdateOperationalProfileRequest,
) (*dto.AdminUserResponse, error) {
	if actorID == targetID {
		return nil, domainuser.ErrCannotModifySelf
	}
	target, err := u.users.GetByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domainuser.ErrEmployeeCodeExists
		}
		return nil, err
	}

	var before any
	var after any
	switch target.Role.ProfileType() {
	case "staff":
		current, getErr := u.staff.GetByUserID(ctx, targetID)
		if getErr != nil {
			return nil, getErr
		}
		employeeCode := current.EmployeeCode
		if req.EmployeeCode != nil {
			employeeCode = *req.EmployeeCode
		}
		hireDate := current.HireDate
		if req.HireDate != nil {
			parsed, parseErr := time.Parse("2006-01-02", *req.HireDate)
			if parseErr != nil {
				return nil, apperrors.ErrValidation.WithDetail("hire_date", "must match YYYY-MM-DD")
			}
			hireDate = parsed
		}
		shift := current.Shift
		if req.Shift != nil {
			shift = *req.Shift
		}
		before = current
		after, err = u.staff.UpdateByUserID(ctx, port.UpsertStaffProfileParams{
			UserID:       targetID,
			FullName:     req.FullName,
			Phone:        req.Phone,
			EmployeeCode: employeeCode,
			HireDate:     hireDate,
			Shift:        shift,
		})
	case "admin":
		current, getErr := u.admins.GetByUserID(ctx, targetID)
		if getErr != nil {
			return nil, getErr
		}
		before = current
		after, err = u.admins.UpdateByUserID(ctx, port.UpsertAdminProfileParams{
			UserID:   targetID,
			FullName: req.FullName,
			Phone:    req.Phone,
		})
	default:
		return nil, domainuser.ErrInvalidRoleTransition
	}
	if err != nil {
		return nil, err
	}

	u.logAudit(ctx, domainuser.AuditActionAdminUpdatedProfile, actorID, actorRole, &targetID, "user_profile", before, after)
	return u.getAdminUserResponse(ctx, targetID)
}

func (u *AdminUserUsecase) UpdateRole(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	targetID uuid.UUID,
	req dto.UpdateUserRoleRequest,
) (*dto.AdminUserResponse, error) {
	if actorID == targetID {
		return nil, domainuser.ErrCannotModifySelf
	}
	newRole, err := parseRole(req.Role)
	if err != nil {
		return nil, err
	}
	if !newRole.CanBeAssigned() {
		return nil, domainuser.ErrInvalidRoleTransition
	}

	var beforeRole, afterRole string
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		target, getErr := u.users.GetByIDForUpdate(txCtx, targetID)
		if getErr != nil {
			return getErr
		}
		beforeRole = string(target.Role)
		afterRole = string(newRole)
		if target.Role == newRole {
			return nil
		}

		switch target.Role.ProfileType() {
		case "customer":
			if delErr := u.customers.DeleteByUserID(txCtx, targetID); delErr != nil && !errors.Is(delErr, apperrors.ErrNotFound) {
				return delErr
			}
		case "staff":
			if delErr := u.staff.DeleteByUserID(txCtx, targetID); delErr != nil && !errors.Is(delErr, apperrors.ErrNotFound) {
				return delErr
			}
		case "admin":
			if delErr := u.admins.DeleteByUserID(txCtx, targetID); delErr != nil && !errors.Is(delErr, apperrors.ErrNotFound) {
				return delErr
			}
		}

		switch newRole.ProfileType() {
		case "customer":
			_, getErr = u.customers.Create(txCtx, port.UpsertCustomerProfileParams{
				UserID: targetID,
			})
		case "staff":
			employeeCode, hireDate, shift, prepErr := parseStaffSeed(req.EmployeeCode, req.HireDate, req.Shift)
			if prepErr != nil {
				return prepErr
			}
			fullName := req.FullName
			if fullName == "" {
				fullName = strings.Split(target.Email, "@")[0]
			}
			_, getErr = u.staff.Create(txCtx, port.UpsertStaffProfileParams{
				UserID:       targetID,
				FullName:     fullName,
				Phone:        req.Phone,
				EmployeeCode: employeeCode,
				HireDate:     hireDate,
				Shift:        shift,
			})
		case "admin":
			fullName := req.FullName
			if fullName == "" {
				fullName = strings.Split(target.Email, "@")[0]
			}
			_, getErr = u.admins.Create(txCtx, port.UpsertAdminProfileParams{
				UserID:   targetID,
				FullName: fullName,
				Phone:    req.Phone,
			})
		default:
			return domainuser.ErrInvalidRoleTransition
		}
		if getErr != nil {
			return getErr
		}
		return u.users.UpdateRole(txCtx, targetID, newRole)
	}); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domainuser.ErrEmployeeCodeExists
		}
		return nil, err
	}

	u.logAudit(ctx, domainuser.AuditActionAdminUpdatedRole, actorID, actorRole, &targetID, "user", map[string]any{"role": beforeRole}, map[string]any{"role": afterRole})
	return u.getAdminUserResponse(ctx, targetID)
}

func (u *AdminUserUsecase) Disable(ctx context.Context, actorID uuid.UUID, actorRole domainuser.Role, targetID uuid.UUID) error {
	if actorID == targetID {
		return domainuser.ErrCannotModifySelf
	}
	if err := u.users.SoftDelete(ctx, targetID); err != nil {
		return err
	}
	if err := u.sessions.DeleteAllForUser(ctx, targetID.String()); err != nil {
		return err
	}
	u.logAudit(ctx, domainuser.AuditActionAdminDisabledUser, actorID, actorRole, &targetID, "user", nil, map[string]any{"disabled": true})
	return nil
}

func (u *AdminUserUsecase) RevokeSessions(ctx context.Context, actorID uuid.UUID, actorRole domainuser.Role, targetID uuid.UUID) error {
	if actorID == targetID {
		return domainuser.ErrCannotModifySelf
	}
	if err := u.sessions.DeleteAllForUser(ctx, targetID.String()); err != nil {
		return err
	}
	u.logAudit(ctx, domainuser.AuditActionAdminRevokedUserSessions, actorID, actorRole, &targetID, "session", nil, map[string]any{"revoked": true})
	return nil
}

func (u *AdminUserUsecase) Get(ctx context.Context, userID uuid.UUID) (*dto.AdminUserResponse, error) {
	return u.getAdminUserResponse(ctx, userID)
}

func (u *AdminUserUsecase) List(ctx context.Context, page, pageSize int, search string) ([]dto.AdminUserResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	off := int64(page-1) * int64(pageSize)
	if off > math.MaxInt32 {
		off = math.MaxInt32
	}
	items, total, err := u.users.AdminList(ctx, port.AdminListUsersParams{
		Search: strings.TrimSpace(search),
		Limit:  int32(pageSize),
		Offset: int32(off),
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.AdminUserResponse, 0, len(items))
	for _, item := range items {
		out = append(out, mapAdminListItem(item))
	}
	return out, total, nil
}

func (u *AdminUserUsecase) getAdminUserResponse(ctx context.Context, userID uuid.UUID) (*dto.AdminUserResponse, error) {
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := &dto.AdminUserResponse{
		ID:                 user.ID.String(),
		Email:              user.Email,
		Role:               string(user.Role),
		EmailVerified:      user.EmailVerified,
		MustChangePassword: user.MustChangePassword,
		Disabled:           user.Disabled(),
		CreatedAt:          user.CreatedAt,
		UpdatedAt:          user.UpdatedAt,
	}
	profileAny, err := u.profiles.GetByUserID(ctx, userID, user.Role)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	switch p := profileAny.(type) {
	case *domainprofile.Customer:
		resp.DisplayName = p.DisplayName
		resp.Phone = p.Phone
	case *domainprofile.Staff:
		resp.FullName = &p.FullName
		resp.Phone = p.Phone
		resp.EmployeeCode = &p.EmployeeCode
		t := p.HireDate
		resp.HireDate = &t
		resp.Shift = &p.Shift
	case *domainprofile.Admin:
		resp.FullName = &p.FullName
		resp.Phone = p.Phone
	}
	return resp, nil
}

func mapAdminListItem(in port.AdminListUser) dto.AdminUserResponse {
	return dto.AdminUserResponse{
		ID:                 in.ID.String(),
		Email:              in.Email,
		Role:               string(in.Role),
		EmailVerified:      in.EmailVerified,
		MustChangePassword: in.MustChangePassword,
		Disabled:           in.DeletedAt != nil,
		CreatedAt:          in.CreatedAt,
		UpdatedAt:          in.UpdatedAt,
		DisplayName:        in.DisplayName,
		FullName:           in.FullName,
		Phone:              in.Phone,
		EmployeeCode:       in.EmployeeCode,
		HireDate:           in.HireDate,
		Shift:              in.Shift,
	}
}

func parseRole(raw string) (domainuser.Role, error) {
	role := domainuser.Role(strings.TrimSpace(strings.ToLower(raw)))
	if !role.CanBeAssigned() {
		return "", apperrors.ErrValidation.WithDetail("role", "unsupported role")
	}
	return role, nil
}

func parseStaffSeed(employeeCode, hireDate, shift *string) (string, time.Time, string, error) {
	if employeeCode == nil || strings.TrimSpace(*employeeCode) == "" {
		return "", time.Time{}, "", apperrors.ErrValidation.WithDetail("employee_code", "required for staff roles")
	}
	if hireDate == nil || strings.TrimSpace(*hireDate) == "" {
		return "", time.Time{}, "", apperrors.ErrValidation.WithDetail("hire_date", "required for staff roles")
	}
	parsed, err := time.Parse("2006-01-02", *hireDate)
	if err != nil {
		return "", time.Time{}, "", apperrors.ErrValidation.WithDetail("hire_date", "must match YYYY-MM-DD")
	}
	if shift == nil || strings.TrimSpace(*shift) == "" {
		return "", time.Time{}, "", apperrors.ErrValidation.WithDetail("shift", "required for staff roles")
	}
	return strings.TrimSpace(*employeeCode), parsed, strings.TrimSpace(*shift), nil
}

func (u *AdminUserUsecase) logAudit(ctx context.Context, action domainuser.AuditAction, actorID uuid.UUID, actorRole domainuser.Role, targetID *uuid.UUID, targetType string, before, after any) {
	if u.audit == nil {
		return
	}
	_ = u.audit.Log(ctx, action, actorID, actorRole, targetID, targetType, before, after)
}
