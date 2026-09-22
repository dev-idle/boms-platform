package usecase

import (
	"context"
	"errors"
	"strings"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/service/profilesvc"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AdminUserUsecase struct {
	users     port.UserRepository
	staff     port.StaffProfileRepository
	sessions  port.SessionStore
	tx        port.TxManager
	hasher    port.PasswordHasher
	audit     *auditlogger.Service
	auditLogs port.AuditLogRepository
	profiles  *profilesvc.Service
	log       *zap.Logger
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
	auditLogs port.AuditLogRepository,
	log *zap.Logger,
) *AdminUserUsecase {
	return &AdminUserUsecase{
		users:     users,
		staff:     staff,
		sessions:  sessions,
		tx:        tx,
		hasher:    hasher,
		audit:     audit,
		auditLogs: auditLogs,
		profiles:  profilesvc.NewService(customers, staff, admins),
		log:       log,
	}
}

func (u *AdminUserUsecase) CreateOperationalUser(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.CreateOperationalUserRequest,
) (*dto.CreateOperationalUserResponse, error) {
	phone, err := normalizeRequestPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	// On create there is nothing to clear: a blank phone means none.
	phone = resolvePatchString(phone, nil)
	role, err := parseRole(req.Role)
	if err != nil {
		return nil, err
	}
	tempPassword, err := utils.GenerateTempPassword(16)
	if err != nil {
		return nil, err
	}
	hash, err := u.hasher.Hash(tempPassword)
	if err != nil {
		return nil, apperrors.Errorf("hash temp password: %w", err)
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
			if errors.Is(createErr, apperrors.ErrConflict) {
				return ErrEmailExists
			}
			return createErr
		}
		user = created
		if claimErr := claimPhone(txCtx, u.users, user.ID, phone, nil); claimErr != nil {
			return claimErr
		}
		return u.createStaffProfile(txCtx, user.ID, req.FullName, phone, req.EmployeeCode)
	}); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domainuser.ErrEmployeeCodeExists
		}
		return nil, err
	}

	created, err := u.getAdminUserResponse(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	u.logAudit(ctx, domainuser.AuditActionAdminCreatedUser, actorID, actorRole, &user.ID, "user", nil, map[string]any{"role": req.Role})
	return &dto.CreateOperationalUserResponse{
		User:         *created,
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
	phone, err := normalizeRequestPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	// The DTO's `required` tag passes a blank string; this endpoint sets the name,
	// so a blank one is an error rather than "keep the current name".
	if strings.TrimSpace(req.FullName) == "" {
		return nil, apperrors.ErrValidation.WithDetail("full_name", "required")
	}
	if actorID == targetID {
		return nil, domainuser.ErrCannotModifySelf
	}

	var before, after *domainprofile.Staff
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, lockErr := u.lockStaffTarget(txCtx, targetID); lockErr != nil {
			return lockErr
		}
		var writeErr error
		before, after, writeErr = u.patchStaffProfile(txCtx, targetID, staffProfilePatch{
			FullName:     req.FullName,
			Phone:        phone,
			EmployeeCode: req.EmployeeCode,
		})
		return writeErr
	}); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domainuser.ErrEmployeeCodeExists
		}
		return nil, err
	}

	u.logAudit(ctx, domainuser.AuditActionAdminUpdatedProfile, actorID, actorRole, &targetID, "user_profile", before, after)
	return u.getAdminUserResponse(ctx, targetID)
}

// UpdateRole moves a staff, baker or manager account to another of those roles
// and saves the name and phone sent with it. The three roles share one staff
// profile, so the profile is edited in place and the employee code stays with the
// person; only the role itself changes, and only a real change revokes sessions.
func (u *AdminUserUsecase) UpdateRole(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	targetID uuid.UUID,
	req dto.UpdateUserRoleRequest,
) (*dto.AdminUserResponse, error) {
	phone, err := normalizeRequestPhone(req.Phone)
	if err != nil {
		return nil, err
	}
	if actorID == targetID {
		return nil, domainuser.ErrCannotModifySelf
	}
	newRole, err := parseRole(req.Role)
	if err != nil {
		return nil, err
	}

	var change roleChange
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		var applyErr error
		change, applyErr = u.applyRoleChange(txCtx, targetID, newRole, staffProfilePatch{
			FullName:     req.FullName,
			Phone:        phone,
			EmployeeCode: req.EmployeeCode,
		})
		return applyErr
	}); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domainuser.ErrEmployeeCodeExists
		}
		return nil, err
	}

	u.logAudit(ctx, domainuser.AuditActionAdminUpdatedProfile, actorID, actorRole, &targetID, "user_profile", change.before, change.after)
	if change.oldRole != newRole {
		if err := u.sessions.DeleteAllForUser(ctx, targetID.String()); err != nil {
			return nil, err
		}
		u.logAudit(ctx, domainuser.AuditActionAdminUpdatedRole, actorID, actorRole, &targetID, "user", map[string]any{"role": string(change.oldRole)}, map[string]any{"role": string(newRole)})
	}
	return u.getAdminUserResponse(ctx, targetID)
}

// roleChange is what UpdateRole's transaction reports back for the audit trail.
type roleChange struct {
	oldRole       domainuser.Role
	before, after *domainprofile.Staff
}

// applyRoleChange runs inside the caller's transaction. The target row is locked
// first, so two admins changing one account serialize instead of both reading the
// same old role and overwriting each other's profile edits.
func (u *AdminUserUsecase) applyRoleChange(
	ctx context.Context,
	targetID uuid.UUID,
	newRole domainuser.Role,
	patch staffProfilePatch,
) (roleChange, error) {
	target, err := u.lockStaffTarget(ctx, targetID)
	if err != nil {
		return roleChange{}, err
	}
	before, after, err := u.patchStaffProfile(ctx, targetID, patch)
	if err != nil {
		return roleChange{}, err
	}
	change := roleChange{oldRole: target.Role, before: before, after: after}
	if newRole == target.Role {
		return change, nil
	}
	return change, u.users.UpdateRole(ctx, targetID, newRole)
}

// lockStaffTarget locks the account an admin is editing, inside the caller's
// transaction, and checks it is an active staff, baker or manager account.
func (u *AdminUserUsecase) lockStaffTarget(ctx context.Context, targetID uuid.UUID) (*domainuser.User, error) {
	target, err := u.users.AdminGetByIDForUpdate(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if err := refuseAdminTarget(target); err != nil {
		return nil, err
	}
	if target.Disabled() {
		return nil, apperrors.ErrValidation.WithDetail("account", "user is disabled")
	}
	if !isStaffProfileRole(target.Role) {
		return nil, domainuser.ErrInvalidRoleTransition
	}
	return target, nil
}

// staffProfilePatch carries the admin-editable staff fields with PATCH meaning:
// an empty name or a nil phone or code keeps the stored value; "" clears a phone.
type staffProfilePatch struct {
	FullName     string
	Phone        *string
	EmployeeCode *string
}

func (u *AdminUserUsecase) patchStaffProfile(
	ctx context.Context,
	userID uuid.UUID,
	patch staffProfilePatch,
) (before, after *domainprofile.Staff, err error) {
	current, err := u.staff.GetByUserID(ctx, userID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, nil, domainuser.ErrProfileNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	phone := resolvePatchString(patch.Phone, current.Phone)
	if err := claimPhone(ctx, u.users, userID, phone, current.Phone); err != nil {
		return nil, nil, err
	}
	fullName := current.FullName
	if name := strings.TrimSpace(patch.FullName); name != "" {
		fullName = name
	}
	employeeCode := current.EmployeeCode
	if code := resolvePatchString(patch.EmployeeCode, nil); code != nil {
		employeeCode = *code
	}
	updated, err := u.staff.UpdateByUserID(ctx, port.UpsertStaffProfileParams{
		UserID:       userID,
		FullName:     fullName,
		Phone:        phone,
		EmployeeCode: employeeCode,
	})
	if err != nil {
		return nil, nil, err
	}
	return current, updated, nil
}

func (u *AdminUserUsecase) Enable(ctx context.Context, actorID uuid.UUID, actorRole domainuser.Role, targetID uuid.UUID) error {
	if actorID == targetID {
		return domainuser.ErrCannotModifySelf
	}
	target, err := u.mutableTarget(ctx, targetID)
	if err != nil {
		return err
	}
	var released *string
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if restoreErr := u.users.Restore(txCtx, targetID); restoreErr != nil {
			return restoreErr
		}
		var reclaimErr error
		released, reclaimErr = u.reclaimPhone(txCtx, targetID, target.Role)
		return reclaimErr
	}); err != nil {
		return err
	}
	after := map[string]any{"disabled": false}
	if released != nil {
		after["phone_released"] = *released
	}
	u.logAudit(ctx, domainuser.AuditActionAdminEnabledUser, actorID, actorRole, &targetID, "user", map[string]any{"disabled": true}, after)
	return nil
}

// reclaimPhone re-checks the phone of an account coming back from a soft delete.
// While it was disabled its number counted as free. If an active account has
// taken it since, that holder keeps it and the returning account comes back
// without a phone; the enable audit entry records the number released. Refusing
// the enable instead would strand the account: every admin edit path refuses
// disabled users, and customers have no admin edit path at all.
func (u *AdminUserUsecase) reclaimPhone(ctx context.Context, userID uuid.UUID, role domainuser.Role) (released *string, err error) {
	profile, err := u.profiles.GetByUserID(ctx, userID, role)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	phone := profilePhone(profile)
	claimErr := claimPhone(ctx, u.users, userID, phone, nil)
	if !errors.Is(claimErr, domainuser.ErrPhoneExists) {
		return nil, claimErr
	}
	if err := u.users.ReleasePhone(ctx, userID); err != nil {
		return nil, err
	}
	return phone, nil
}

func (u *AdminUserUsecase) ResetPassword(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	targetID uuid.UUID,
) (*dto.AdminResetPasswordResponse, error) {
	if actorID == targetID {
		return nil, domainuser.ErrCannotModifySelf
	}
	// The response carries the new password; resetting another admin would hand
	// over that account.
	target, err := u.mutableTarget(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if target.Disabled() {
		return nil, apperrors.ErrValidation.WithDetail("account", "user is disabled")
	}
	if !target.Role.IsOperational() {
		return nil, domainuser.ErrInvalidRoleTransition
	}

	tempPassword, err := utils.GenerateTempPassword(16)
	if err != nil {
		return nil, err
	}
	hash, err := u.hasher.Hash(tempPassword)
	if err != nil {
		return nil, apperrors.Errorf("hash temp password: %w", err)
	}
	if err := u.users.AdminUpdatePassword(ctx, targetID, hash); err != nil {
		return nil, err
	}
	if err := u.sessions.DeleteAllForUser(ctx, targetID.String()); err != nil {
		return nil, err
	}

	u.logAudit(ctx, domainuser.AuditActionAdminResetUserPassword, actorID, actorRole, &targetID, "user", nil, map[string]any{"must_change_password": true})
	resp, err := u.getAdminUserResponse(ctx, targetID)
	if err != nil {
		return nil, err
	}
	return &dto.AdminResetPasswordResponse{
		User:         *resp,
		TempPassword: tempPassword,
	}, nil
}

func (u *AdminUserUsecase) Disable(ctx context.Context, actorID uuid.UUID, actorRole domainuser.Role, targetID uuid.UUID) error {
	if actorID == targetID {
		return domainuser.ErrCannotModifySelf
	}
	if _, err := u.mutableTarget(ctx, targetID); err != nil {
		return err
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
	if _, err := u.mutableTarget(ctx, targetID); err != nil {
		return err
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

func (u *AdminUserUsecase) PreviewNextEmployeeCode(ctx context.Context) (*dto.NextEmployeeCodeResponse, error) {
	code, err := u.staff.NextEmployeeCode(ctx)
	if err != nil {
		return nil, err
	}
	return &dto.NextEmployeeCodeResponse{EmployeeCode: code}, nil
}

func (u *AdminUserUsecase) List(
	ctx context.Context,
	rawPage, rawPageSize int32,
	search string,
	roleFilter string,
) (items []dto.AdminUserResponse, total int64, page, pageSize int32, err error) {
	page, pageSize = normalizeAdminUserListPage(rawPage, rawPageSize)
	role, err := parseAdminListRoleFilter(roleFilter)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	rows, total, err := u.users.AdminList(ctx, port.AdminListUsersParams{
		Search: strings.TrimSpace(search),
		Role:   role,
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, 0, 0, err
	}
	out := make([]dto.AdminUserResponse, 0, len(rows))
	for _, item := range rows {
		out = append(out, mapAdminListItem(item))
	}
	return out, total, page, pageSize, nil
}

func (u *AdminUserUsecase) getAdminUserResponse(ctx context.Context, userID uuid.UUID) (*dto.AdminUserResponse, error) {
	user, err := u.users.AdminGetByID(ctx, userID)
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
	}
}

// parseRole accepts the roles admin endpoints may assign: staff, baker, manager.
func parseRole(raw string) (domainuser.Role, error) {
	role := domainuser.Role(strings.TrimSpace(strings.ToLower(raw)))
	if !isStaffProfileRole(role) {
		return "", apperrors.ErrValidation.WithDetail("role", "unsupported role")
	}
	return role, nil
}

// isStaffProfileRole reports the roles backed by the shared staff profile.
func isStaffProfileRole(role domainuser.Role) bool {
	return role.ProfileType() == "staff"
}

// mutableTarget loads the account an admin endpoint is about to change and refuses
// admin accounts: admins are managed through the dev seed, never by each other.
func (u *AdminUserUsecase) mutableTarget(ctx context.Context, targetID uuid.UUID) (*domainuser.User, error) {
	target, err := u.users.AdminGetByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if err := refuseAdminTarget(target); err != nil {
		return nil, err
	}
	return target, nil
}

func refuseAdminTarget(target *domainuser.User) error {
	if target.Role.IsAdmin() {
		return domainuser.ErrCannotModifyAdmin
	}
	return nil
}

func parseAdminListRoleFilter(raw string) (*domainuser.Role, error) {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return nil, nil
	}
	role := domainuser.Role(trimmed)
	switch role {
	case domainuser.RoleCustomer, domainuser.RoleStaff, domainuser.RoleBaker, domainuser.RoleManager, domainuser.RoleAdmin:
		return &role, nil
	default:
		return nil, apperrors.ErrValidation.WithDetail("role", "unsupported role")
	}
}

func (u *AdminUserUsecase) createStaffProfile(
	ctx context.Context,
	userID uuid.UUID,
	fullName string,
	phone *string,
	requestedCode *string,
) error {
	// No retry on conflict: a failed statement aborts the surrounding transaction, and
	// NextEmployeeCode already serializes allocation with a transaction-scoped advisory lock.
	employeeCode, err := u.resolveStaffEmployeeCode(ctx, requestedCode)
	if err != nil {
		return err
	}
	_, err = u.staff.Create(ctx, port.UpsertStaffProfileParams{
		UserID:       userID,
		FullName:     fullName,
		Phone:        phone,
		EmployeeCode: employeeCode,
	})
	return err
}

func (u *AdminUserUsecase) resolveStaffEmployeeCode(ctx context.Context, requested *string) (string, error) {
	if requested != nil {
		code := strings.TrimSpace(*requested)
		if code != "" {
			return code, nil
		}
	}
	return u.staff.NextEmployeeCode(ctx)
}

func (u *AdminUserUsecase) logAudit(ctx context.Context, action domainuser.AuditAction, actorID uuid.UUID, actorRole domainuser.Role, targetID *uuid.UUID, targetType string, before, after any) {
	recordAudit(u.log, u.audit, ctx, action, actorID, actorRole, targetID, targetType, before, after)
}
