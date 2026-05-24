package middleware

import (
	"context"
	"errors"
	"strings"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	localUserIDKey    = "auth_user_id"
	localRoleKey      = "auth_role"
	localSessionIDKey = "auth_session_id"
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
		locals, ok := parseAccessToken(c, signer)
		if !ok {
			return unauthorized(c)
		}
		setAuthLocals(c, locals)
		return c.Next()
	}
}

// RequireAuthWithSession parses Bearer access JWT and verifies the session exists in Redis.
// Use for state-changing or security-sensitive routes (orders, payments, admin writes).
func RequireAuthWithSession(signer port.TokenSigner, sessions port.SessionStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		locals, ok := parseAccessToken(c, signer)
		if !ok {
			return unauthorized(c)
		}
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		if _, err := sessions.Get(ctx, locals.UserID.String(), locals.SessionID.String()); err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return sessionRevoked(c)
			}
			return response.Error(c, fiber.StatusInternalServerError, &response.ErrorBody{
				Code:    apperrors.ErrInternal.Code,
				Message: apperrors.ErrInternal.Message,
			})
		}
		setAuthLocals(c, locals)
		return c.Next()
	}
}

// OptionalAuth parses Bearer token when present; missing or invalid token is allowed.
func OptionalAuth(signer port.TokenSigner) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if locals, ok := parseAccessToken(c, signer); ok {
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

func parseAccessToken(c *fiber.Ctx, signer port.TokenSigner) (authLocals, bool) {
	token, ok := bearerToken(c)
	if !ok {
		return authLocals{}, false
	}
	claims, err := signer.ParseAccess(token)
	if err != nil {
		return authLocals{}, false
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil || userID == uuid.Nil {
		return authLocals{}, false
	}
	sessionID, err := uuid.Parse(claims.SessionID)
	if err != nil || sessionID == uuid.Nil {
		return authLocals{}, false
	}
	if claims.Role == "" {
		return authLocals{}, false
	}
	return authLocals{
		UserID:    userID,
		Role:      domainuser.Role(claims.Role),
		SessionID: sessionID,
	}, true
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

func unauthorized(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusUnauthorized, &response.ErrorBody{
		Code:    "UNAUTHORIZED",
		Message: "Authentication required",
	})
}

func sessionRevoked(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusUnauthorized, &response.ErrorBody{
		Code:    "SESSION_REVOKED",
		Message: "Session revoked",
	})
}

func forbidden(c *fiber.Ctx) error {
	return response.Error(c, fiber.StatusForbidden, &response.ErrorBody{
		Code:    "FORBIDDEN",
		Message: "Insufficient permissions",
	})
}
