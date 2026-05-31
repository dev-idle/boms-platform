package v1

import (
	"errors"

	"github.com/boms/backend/internal/config"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AuthHandler exposes /api/v1/auth endpoints.
type AuthHandler struct {
	usecase *usecase.AuthUsecase
	cfg     *config.Config
}

// NewAuthHandler wires auth use case and configuration.
func NewAuthHandler(uc *usecase.AuthUsecase, cfg *config.Config) *AuthHandler {
	return &AuthHandler{usecase: uc, cfg: cfg}
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("reason", "invalid_body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	user, err := h.usecase.Register(c.UserContext(), req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, toUserResponse(user))
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("reason", "invalid_body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	access, refresh, user, err := h.usecase.Login(c.UserContext(), req, c.Get(fiber.HeaderUserAgent), c.IP())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	writeRefreshCookie(c, h.cfg, refresh)
	noStore(c)
	return response.OK(c, dto.TokenResponse{
		AccessToken:        access,
		TokenType:          "Bearer",
		ExpiresIn:          int(h.cfg.JWT.AccessTTL.Seconds()),
		User:               toUserResponse(user),
		MustChangePassword: mustChangePasswordPtr(user.MustChangePassword),
	})
}

// Refresh handles POST /api/v1/auth/refresh (reads refresh cookie).
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	refresh := c.Cookies(h.cfg.Cookie.Name)
	if refresh == "" {
		return writeAppError(c, apperrors.ErrMissingRefreshToken)
	}
	access, newRefresh, mustChange, err := h.usecase.Refresh(c.UserContext(), refresh, c.Get(fiber.HeaderUserAgent), c.IP())
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidRefreshToken) || errors.Is(err, apperrors.ErrSessionRevoked) {
			clearRefreshCookie(c, h.cfg)
		}
		return writeMapUsecaseError(c, err)
	}
	writeRefreshCookie(c, h.cfg, newRefresh)
	noStore(c)
	return response.OK(c, dto.RefreshResponse{
		AccessToken:        access,
		TokenType:          "Bearer",
		ExpiresIn:          int(h.cfg.JWT.AccessTTL.Seconds()),
		MustChangePassword: mustChangePasswordPtr(mustChange),
	})
}

// noStore marks responses that carry credentials (access token) as non-cacheable.
// Prevents browsers / reverse proxies from caching bearer tokens.
func noStore(c *fiber.Ctx) {
	c.Set(fiber.HeaderCacheControl, "no-store")
	c.Set(fiber.HeaderPragma, "no-cache")
}

// Logout handles POST /api/v1/auth/logout (hybrid idempotent: Bearer preferred, cookie fallback).
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	bearerUserID, bearerSessionID, hasBearer := bearerSessionFromCtx(c)
	h.usecase.LogoutHybrid(c.UserContext(), bearerUserID, bearerSessionID, hasBearer, c.Cookies(h.cfg.Cookie.Name))
	clearRefreshCookie(c, h.cfg)
	return response.NoContent(c)
}

func toUserResponse(u *domainuser.User) dto.UserResponse {
	if u == nil {
		return dto.UserResponse{}
	}
	return dto.UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Role:          string(u.Role),
		EmailVerified: u.EmailVerified,
		CreatedAt:     u.CreatedAt,
	}
}

func writeRefreshCookie(c *fiber.Ctx, cfg *config.Config, token string) {
	c.Cookie(refreshCookie(cfg, token, int(cfg.JWT.RefreshTTL.Seconds())))
}

func clearRefreshCookie(c *fiber.Ctx, cfg *config.Config) {
	c.Cookie(refreshCookie(cfg, "", -1))
}

func refreshCookie(cfg *config.Config, token string, maxAge int) *fiber.Cookie {
	cookie := &fiber.Cookie{
		Name:     cfg.Cookie.Name,
		Value:    token,
		Path:     middleware.AuthCookiePath,
		HTTPOnly: true,
		Secure:   cfg.Cookie.Secure,
		SameSite: fiber.CookieSameSiteLaxMode,
		MaxAge:   maxAge,
	}
	if cfg.Cookie.Domain != "" {
		cookie.Domain = cfg.Cookie.Domain
	}
	return cookie
}

// bearerSessionFromCtx returns Bearer-derived session identity when OptionalAuth populated Locals.
func bearerSessionFromCtx(c *fiber.Ctx) (userID, sessionID uuid.UUID, ok bool) {
	uid, uidOK := middleware.GetUserID(c)
	sid, sidOK := middleware.GetSessionID(c)
	if !uidOK || !sidOK {
		return uuid.Nil, uuid.Nil, false
	}
	return uid, sid, true
}
