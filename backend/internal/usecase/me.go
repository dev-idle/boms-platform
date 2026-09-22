package usecase

import (
	"context"
	"errors"
	"net/http"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/service/profilesvc"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrMeNotFound = apperrors.New(http.StatusNotFound, "me_not_found", "User not found")
)

type MeUsecase struct {
	users     port.UserRepository
	customers port.CustomerProfileRepository
	staff     port.StaffProfileRepository
	admins    port.AdminProfileRepository
	sessions  port.SessionStore
	tx        port.TxManager
	hasher    port.PasswordHasher
	audit     *auditlogger.Service
	profiles  *profilesvc.Service
	log       *zap.Logger
}

func NewMeUsecase(
	users port.UserRepository,
	customers port.CustomerProfileRepository,
	staff port.StaffProfileRepository,
	admins port.AdminProfileRepository,
	sessions port.SessionStore,
	tx port.TxManager,
	hasher port.PasswordHasher,
	audit *auditlogger.Service,
	log *zap.Logger,
) *MeUsecase {
	return &MeUsecase{
		users:     users,
		customers: customers,
		staff:     staff,
		admins:    admins,
		sessions:  sessions,
		tx:        tx,
		hasher:    hasher,
		audit:     audit,
		profiles:  profilesvc.NewService(customers, staff, admins),
		log:       log,
	}
}

func (u *MeUsecase) Get(ctx context.Context, userID uuid.UUID) (*domainuser.User, any, error) {
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, nil, ErrMeNotFound
		}
		return nil, nil, err
	}
	profile, err := u.profiles.GetByUserID(ctx, user.ID, user.Role)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return user, nil, domainuser.ErrProfileNotFound
		}
		return user, nil, err
	}
	return user, profile, nil
}

func (u *MeUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateMeRequest) (*domainuser.User, any, error) {
	phone, err := normalizeRequestPhone(req.Phone)
	if err != nil {
		return nil, nil, err
	}
	req.Phone = phone
	user, before, err := u.Get(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	var profile any
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		if claimErr := claimPhone(txCtx, u.users, userID, req.Phone, profilePhone(before)); claimErr != nil {
			return claimErr
		}
		var writeErr error
		profile, writeErr = u.writeProfile(txCtx, userID, before, req)
		return writeErr
	}); err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, nil, domainuser.ErrEmployeeCodeExists
		}
		return nil, nil, err
	}

	u.logAudit(ctx, domainuser.AuditActionMeUpdatedProfile, userID, user.Role, &userID, "user_profile", before, profile)
	return user, profile, nil
}

func (u *MeUsecase) writeProfile(ctx context.Context, userID uuid.UUID, current any, req dto.UpdateMeRequest) (any, error) {
	switch p := current.(type) {
	case *domainprofile.Customer:
		return u.customers.UpdateByUserID(ctx, port.UpsertCustomerProfileParams{
			UserID:      userID,
			DisplayName: resolvePatchString(req.DisplayName, p.DisplayName),
			Phone:       resolvePatchString(req.Phone, p.Phone),
		})
	case *domainprofile.Staff:
		fullName := p.FullName
		if req.FullName != nil {
			fullName = *req.FullName
		}
		return u.staff.UpdateByUserID(ctx, port.UpsertStaffProfileParams{
			UserID:       userID,
			FullName:     fullName,
			Phone:        resolvePatchString(req.Phone, p.Phone),
			EmployeeCode: p.EmployeeCode,
		})
	case *domainprofile.Admin:
		fullName := p.FullName
		if req.FullName != nil {
			fullName = *req.FullName
		}
		return u.admins.UpdateByUserID(ctx, port.UpsertAdminProfileParams{
			UserID:   userID,
			FullName: fullName,
			Phone:    resolvePatchString(req.Phone, p.Phone),
		})
	default:
		return nil, domainuser.ErrProfileNotFound
	}
}

func (u *MeUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPwd, newPwd string) error {
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return ErrMeNotFound
		}
		return err
	}
	if err := u.hasher.Verify(user.PasswordHash, oldPwd); err != nil {
		return apperrors.ErrInvalidCredentials
	}
	hash, err := u.hasher.Hash(newPwd)
	if err != nil {
		return apperrors.Errorf("hash password: %w", err)
	}
	if err := u.users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	if err := u.users.ClearMustChangePassword(ctx, userID); err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return err
	}
	if err := u.sessions.DeleteAllForUser(ctx, userID.String()); err != nil {
		return err
	}
	u.logAudit(ctx, domainuser.AuditActionMeChangedPassword, userID, user.Role, &userID, "user", nil, map[string]any{"changed": true})
	return nil
}

func (u *MeUsecase) SoftDeleteSelf(ctx context.Context, userID uuid.UUID) error {
	user, err := u.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return ErrMeNotFound
		}
		return err
	}
	if user.Role != domainuser.RoleCustomer {
		return domainuser.ErrSelfDeleteCustomerOnly
	}
	if err := u.users.SoftDelete(ctx, userID); err != nil {
		return err
	}
	if err := u.sessions.DeleteAllForUser(ctx, userID.String()); err != nil {
		return err
	}
	u.logAudit(ctx, domainuser.AuditActionMeSoftDeleted, userID, user.Role, &userID, "user", map[string]any{"disabled": false}, map[string]any{"disabled": true})
	return nil
}

func (u *MeUsecase) logAudit(ctx context.Context, action domainuser.AuditAction, actorID uuid.UUID, actorRole domainuser.Role, targetID *uuid.UUID, targetType string, before, after any) {
	recordAudit(u.log, u.audit, ctx, action, actorID, actorRole, targetID, targetType, before, after)
}
