package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	domainsession "github.com/boms/backend/internal/domain/session"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TimingSafeDummySeed is hashed once at usecase construction to produce dummyHash for login timing equalization.
// Not a user password and never accepted as valid credentials.
const TimingSafeDummySeed = "boms-timing-equalization-seed-v1"

var (
	// ErrEmailExists is returned when registering a duplicate email.
	ErrEmailExists = apperrors.New(http.StatusConflict, "email_exists", "Email already registered")
	// ErrUserNotFound is returned when an authenticated user no longer exists.
	ErrUserNotFound = apperrors.New(http.StatusNotFound, "user_not_found", "User not found")
)

// AuthUsecase implements registration, login, refresh rotation, and session lifecycle.
type AuthUsecase struct {
	userRepo            port.UserRepository
	customerProfileRepo port.CustomerProfileRepository
	txManager           port.TxManager
	sessionStore        port.SessionStore
	hasher              port.PasswordHasher
	signer              port.TokenSigner
	log                 *zap.Logger
	dummyHash           string
}

// NewAuthUsecase wires auth dependencies and precomputes a timing-safe dummy password hash
// using the configured Argon2 parameters (once per process).
func NewAuthUsecase(
	userRepo port.UserRepository,
	customerProfileRepo port.CustomerProfileRepository,
	txManager port.TxManager,
	sessionStore port.SessionStore,
	hasher port.PasswordHasher,
	signer port.TokenSigner,
	log *zap.Logger,
) (*AuthUsecase, error) {
	dummyHash, err := hasher.Hash(TimingSafeDummySeed)
	if err != nil {
		return nil, fmt.Errorf("init timing-safe dummy hash: %w", err)
	}
	return &AuthUsecase{
		userRepo:            userRepo,
		customerProfileRepo: customerProfileRepo,
		txManager:           txManager,
		sessionStore:        sessionStore,
		hasher:              hasher,
		signer:              signer,
		log:                 log,
		dummyHash:           dummyHash,
	}, nil
}

// Register hashes the password and creates a customer account.
func (a *AuthUsecase) Register(ctx context.Context, req dto.RegisterRequest) (*domainuser.User, error) {
	req.Email = utils.NormalizeEmail(req.Email)
	hash, err := a.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	var user *domainuser.User
	runWithTx := func(txCtx context.Context) error {
		var createErr error
		user, createErr = a.userRepo.Create(txCtx, port.CreateUserParams{
			Email:              req.Email,
			PasswordHash:       hash,
			Role:               domainuser.RoleCustomer,
			MustChangePassword: false,
		})
		if createErr != nil {
			return createErr
		}
		if a.customerProfileRepo == nil {
			return fmt.Errorf("customer profile repository is required")
		}
		_, createErr = a.customerProfileRepo.Create(txCtx, port.UpsertCustomerProfileParams{
			UserID: user.ID,
		})
		return createErr
	}
	if a.txManager != nil {
		err = a.txManager.WithTx(ctx, runWithTx)
	} else {
		err = runWithTx(ctx)
	}
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

// Login validates credentials with a single timing-safe Verify, creates a session, and returns JWTs.
func (a *AuthUsecase) Login(ctx context.Context, req dto.LoginRequest, userAgent, ip string) (accessToken, refreshToken string, user *domainuser.User, err error) {
	req.Email = utils.NormalizeEmail(req.Email)
	user, err = a.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return "", "", nil, fmt.Errorf("get user by email: %w", err)
	}

	hashToCheck := a.dummyHash
	if user != nil && !errors.Is(err, apperrors.ErrNotFound) && user.DeletedAt == nil {
		hashToCheck = user.PasswordHash
	}
	if verifyErr := a.hasher.Verify(hashToCheck, req.Password); verifyErr != nil {
		return "", "", nil, apperrors.ErrInvalidCredentials
	}
	if user == nil || errors.Is(err, apperrors.ErrNotFound) || user.DeletedAt != nil {
		return "", "", nil, apperrors.ErrInvalidCredentials
	}

	sid := uuid.NewString()
	jti := uuid.NewString()
	now := time.Now().UTC()
	if err := a.sessionStore.Create(ctx, user.ID.String(), sid, domainsession.SessionMeta{
		RefreshJTI:         jti,
		CreatedAt:          now,
		UserAgent:          userAgent,
		IP:                 ip,
		MustChangePassword: user.MustChangePassword,
	}); err != nil {
		return "", "", nil, fmt.Errorf("create session: %w", err)
	}

	accessToken, err = a.signer.SignAccess(port.AccessTokenClaims{
		Subject:   user.ID.String(),
		Role:      string(user.Role),
		SessionID: sid,
		JTI:       uuid.NewString(),
	})
	if err != nil {
		return "", "", nil, fmt.Errorf("sign access: %w", err)
	}
	refreshToken, err = a.signer.SignRefresh(port.RefreshTokenClaims{
		Subject:   user.ID.String(),
		SessionID: sid,
		JTI:       jti,
	})
	if err != nil {
		return "", "", nil, fmt.Errorf("sign refresh: %w", err)
	}

	if a.log != nil {
		a.log.Info("auth_login_success", zap.String("user_id", user.ID.String()))
	}
	return accessToken, refreshToken, user, nil
}

// Refresh validates the refresh JWT, rotates the session atomically, and issues a new token pair.
func (a *AuthUsecase) Refresh(ctx context.Context, refreshToken, userAgent, ip string) (accessToken, newRefreshToken string, mustChangePassword bool, err error) {
	claims, err := a.signer.ParseRefresh(refreshToken)
	if err != nil {
		return "", "", false, apperrors.ErrInvalidRefreshToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		_ = a.sessionStore.DeleteAllForUser(ctx, claims.Subject)
		return "", "", false, apperrors.ErrSessionRevoked
	}
	// Fetch user before rotation so the new session reflects the latest must_change_password flag.
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			_ = a.sessionStore.DeleteAllForUser(ctx, claims.Subject)
			return "", "", false, apperrors.ErrSessionRevoked
		}
		return "", "", false, apperrors.Errorf("get user: %w", err)
	}

	newSid := uuid.NewString()
	newJti := uuid.NewString()
	now := time.Now().UTC()
	rotateErr := a.sessionStore.Rotate(ctx, claims.Subject, claims.SessionID, newSid, claims.JTI, domainsession.SessionMeta{
		RefreshJTI:         newJti,
		CreatedAt:          now,
		UserAgent:          userAgent,
		IP:                 ip,
		MustChangePassword: user.MustChangePassword,
	})
	if rotateErr != nil {
		if errors.Is(rotateErr, apperrors.ErrNotFound) || errors.Is(rotateErr, apperrors.ErrConflict) {
			_ = a.sessionStore.DeleteAllForUser(ctx, claims.Subject)
			return "", "", false, apperrors.ErrSessionRevoked
		}
		return "", "", false, apperrors.Errorf("rotate session: %w", rotateErr)
	}

	accessToken, err = a.signer.SignAccess(port.AccessTokenClaims{
		Subject:   user.ID.String(),
		Role:      string(user.Role),
		SessionID: newSid,
		JTI:       uuid.NewString(),
	})
	if err != nil {
		return "", "", false, apperrors.Errorf("sign access: %w", err)
	}
	newRefreshToken, err = a.signer.SignRefresh(port.RefreshTokenClaims{
		Subject:   user.ID.String(),
		SessionID: newSid,
		JTI:       newJti,
	})
	if err != nil {
		return "", "", false, apperrors.Errorf("sign refresh: %w", err)
	}

	if a.log != nil {
		a.log.Info("auth_refresh_success", zap.String("user_id", user.ID.String()))
	}
	return accessToken, newRefreshToken, user.MustChangePassword, nil
}

// LogoutSource identifies how a logout request was authenticated.
type LogoutSource string

const (
	// LogoutSourceBearer indicates session revocation via Bearer access token claims.
	LogoutSourceBearer LogoutSource = "bearer"
	// LogoutSourceCookie indicates session revocation via refresh cookie parse.
	LogoutSourceCookie LogoutSource = "cookie"
	// LogoutSourceNone indicates no session could be identified (cookie still cleared).
	LogoutSourceNone LogoutSource = "none"
)

// Logout removes a single session (idempotent — missing key is OK).
func (a *AuthUsecase) Logout(ctx context.Context, userID, sessionID string) error {
	if err := a.sessionStore.Delete(ctx, userID, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// LogoutHybrid best-effort idempotent logout: Bearer path takes priority over refresh cookie.
// Always safe to call; never returns an error (handler responds 204 regardless).
func (a *AuthUsecase) LogoutHybrid(
	ctx context.Context,
	bearerUserID, bearerSessionID uuid.UUID,
	hasBearer bool,
	refreshCookie string,
) LogoutSource {
	if hasBearer && bearerUserID != uuid.Nil && bearerSessionID != uuid.Nil {
		a.revokeSessionBestEffort(ctx, bearerUserID.String(), bearerSessionID.String(), LogoutSourceBearer)
		return LogoutSourceBearer
	}
	if refreshCookie != "" {
		claims, err := a.signer.ParseRefresh(refreshCookie)
		if err == nil {
			userID, uErr := uuid.Parse(claims.Subject)
			sessionID, sErr := uuid.Parse(claims.SessionID)
			if uErr == nil && sErr == nil {
				a.revokeSessionBestEffort(ctx, userID.String(), sessionID.String(), LogoutSourceCookie)
				return LogoutSourceCookie
			}
		}
	}
	if a.log != nil {
		a.log.Info("auth_logout", zap.String("auth_source", string(LogoutSourceNone)))
	}
	return LogoutSourceNone
}

func (a *AuthUsecase) revokeSessionBestEffort(ctx context.Context, userID, sessionID string, source LogoutSource) {
	_ = a.sessionStore.Delete(ctx, userID, sessionID)
	if a.log != nil {
		a.log.Info("auth_logout",
			zap.String("user_id", userID),
			zap.String("session_id", sessionID),
			zap.String("auth_source", string(source)),
		)
	}
}

// ChangePassword verifies the old password, updates the hash, and revokes all sessions.
func (a *AuthUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPwd, newPwd string) error {
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("get user: %w", err)
	}
	if err := a.hasher.Verify(user.PasswordHash, oldPwd); err != nil {
		return apperrors.ErrInvalidCredentials
	}
	hash, err := a.hasher.Hash(newPwd)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := a.userRepo.UpdatePassword(ctx, userID, hash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if err := a.userRepo.ClearMustChangePassword(ctx, userID); err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return fmt.Errorf("clear must-change-password: %w", err)
	}
	if err := a.sessionStore.DeleteAllForUser(ctx, userID.String()); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	return nil
}
