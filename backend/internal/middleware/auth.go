package middleware

import (
	"context"
	"errors"
	"strings"

	domainsession "github.com/boms/backend/internal/domain/session"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	localUserIDKey      = "auth_user_id"
	localRoleKey        = "auth_role"
	localSessionIDKey   = "auth_session_id"
	localSessionMetaKey = "auth_session_meta"
)

// AuthCookiePath is the Path attribute for refresh token cookies.
const AuthCookiePath = "/api/v1/auth"

type authLocals struct {
	UserID    uuid.UUID
	Role      domainuser.Role
	SessionID uuid.UUID
}

// RequireAuth parses Bearer access JWT (stateless) and injects auth locals.
// Does not consult Redis — suitable for read-only routes (e.g. GET /me).
func RequireAuth(signer port.TokenSigner) fiber.Handler {
	return func(c *fiber.Ctx) error {
		locals, authErr := parseAccessToken(c, signer)
		if authErr != nil {
			return writeMiddlewareError(c, authErr)
		}
		setAuthLocals(c, locals)
		return c.Next()
	}
}

// RequireAuthWithSession parses Bearer access JWT and verifies the session exists in Redis.
// Caches the session meta in Fiber Locals so downstream middleware (e.g. RequirePasswordChanged)
// can avoid a second Redis/DB lookup.
func RequireAuthWithSession(signer port.TokenSigner, sessions port.SessionStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		locals, authErr := parseAccessToken(c, signer)
		if authErr != nil {
			return writeMiddlewareError(c, authErr)
		}
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		meta, err := sessions.Get(ctx, locals.UserID.String(), locals.SessionID.String())
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return sessionRevoked(c)
			}
			return writeMiddlewareError(c, apperrors.ErrInternal)
		}
		setAuthLocals(c, locals)
		c.Locals(localSessionMetaKey, meta)
		return c.Next()
	}
}

// OptionalAuth parses Bearer token when present; missing or invalid token is allowed.
func OptionalAuth(signer port.TokenSigner) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if locals, authErr := parseAccessToken(c, signer); authErr == nil {
			setAuthLocals(c, locals)
		}
		return c.Next()
	}
}

// RequireRole ensures the authenticated user has one of the allowed roles.
// Chain after RequireAuth or RequireAuthWithSession.
func RequireRole(roles ...domainuser.Role) fiber.Handler {
	allowed := make(map[domainuser.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *fiber.Ctx) error {
		role, ok := GetRole(c)
		if !ok {
			return forbidden(c)
		}
		if _, ok := allowed[role]; !ok {
			return forbidden(c)
		}
		return c.Next()
	}
}

// GetUserID returns the authenticated user id from Locals.
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	v, ok := c.Locals(localUserIDKey).(uuid.UUID)
	return v, ok && v != uuid.Nil
}

// GetRole returns the authenticated role from Locals.
func GetRole(c *fiber.Ctx) (domainuser.Role, bool) {
	v, ok := c.Locals(localRoleKey).(domainuser.Role)
	return v, ok && v != ""
}

// GetSessionID returns the authenticated session id from Locals.
func GetSessionID(c *fiber.Ctx) (uuid.UUID, bool) {
	v, ok := c.Locals(localSessionIDKey).(uuid.UUID)
	return v, ok && v != uuid.Nil
}

// parseAccessToken returns the populated locals on success or a typed AppError on failure.
// Callers translate the AppError directly so token_expired can be distinguished from generic unauthorized.
func parseAccessToken(c *fiber.Ctx, signer port.TokenSigner) (authLocals, *apperrors.AppError) {
	token, ok := bearerToken(c)
	if !ok {
		return authLocals{}, apperrors.ErrUnauthorized
	}
	claims, err := signer.ParseAccess(token)
	if err != nil {
		if ae, ok := apperrors.AsAppError(err); ok {
			return authLocals{}, ae
		}
		return authLocals{}, apperrors.ErrUnauthorized
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil || userID == uuid.Nil {
		return authLocals{}, apperrors.ErrUnauthorized
	}
	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil || sessionID == uuid.Nil {
		return authLocals{}, apperrors.ErrUnauthorized
	}
	if claims.Role == "" {
		return authLocals{}, apperrors.ErrUnauthorized
	}
	return authLocals{
		UserID:    userID,
		Role:      domainuser.Role(claims.Role),
		SessionID: sessionID,
	}, nil
}

// GetSessionMeta returns the session meta loaded by RequireAuthWithSession, if any.
func GetSessionMeta(c *fiber.Ctx) (domainsession.SessionMeta, bool) {
	v, ok := c.Locals(localSessionMetaKey).(domainsession.SessionMeta)
	return v, ok
}

func setAuthLocals(c *fiber.Ctx, locals authLocals) {
	c.Locals(localUserIDKey, locals.UserID)
	c.Locals(localRoleKey, locals.Role)
	c.Locals(localSessionIDKey, locals.SessionID)
}

func bearerToken(c *fiber.Ctx) (string, bool) {
	auth := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
	if auth == "" {
		return "", false
	}
	const prefix = "bearer "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(auth[len(prefix):])
	return token, token != ""
}

func sessionRevoked(c *fiber.Ctx) error {
	return writeMiddlewareError(c, apperrors.ErrSessionRevoked)
}

func forbidden(c *fiber.Ctx) error {
	return writeMiddlewareError(c, apperrors.ErrForbidden)
}

func writeMiddlewareError(c *fiber.Ctx, e *apperrors.AppError) error {
	code, message, details := e.ToErrorBody()
	return response.Error(c, e.StatusCode, &response.ErrorBody{
		Code:    code,
		Message: message,
		Details: details,
	})
}

// RequirePasswordChanged blocks authenticated users who must change their password.
//
// Resolution order (zero DB hits on the hot path):
//  1. Session meta already loaded by RequireAuthWithSession (zero extra I/O).
//  2. Session store lookup (single Redis GET) — used after RequireAuth-only chains.
//  3. Falls through to allow when no session context is available; route still requires auth.
//
// The flag is set at session creation/rotation and reset when ChangePassword revokes all sessions,
// so the cached value is always at most one refresh cycle stale.
func RequirePasswordChanged(sessions port.SessionStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := GetUserID(c)
		if !ok {
			return c.Next()
		}
		if meta, ok := GetSessionMeta(c); ok {
			if meta.MustChangePassword {
				return writeMiddlewareError(c, apperrors.ErrPasswordChangeRequired)
			}
			return c.Next()
		}
		// Invariant: mutating routes use RequireAuthWithSession so session meta is in Locals.
		// This branch only runs for mis-wired routes; allow through rather than block reads without Redis.
		sessionID, hasSession := GetSessionID(c)
		if !hasSession || sessions == nil {
			return c.Next()
		}
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		meta, err := sessions.Get(ctx, userID.String(), sessionID.String())
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return sessionRevoked(c)
			}
			return writeMiddlewareError(c, apperrors.ErrInternal)
		}
		c.Locals(localSessionMetaKey, meta)
		if meta.MustChangePassword {
			return writeMiddlewareError(c, apperrors.ErrPasswordChangeRequired)
		}
		return c.Next()
	}
}
