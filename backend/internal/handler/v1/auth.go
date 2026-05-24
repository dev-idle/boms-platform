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
		return response.Error(c, fiber.StatusUnprocessableEntity, &response.ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request body",
		})
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, &response.ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: sharevalidator.FieldErrors(err),
		})
	}
	user, err := h.usecase.Register(c.UserContext(), req)
	if err != nil {
		if errors.Is(err, usecase.ErrEmailExists) {
			return response.Error(c, fiber.StatusConflict, &response.ErrorBody{
				Code:    "EMAIL_EXISTS",
				Message: "Email already registered",
			})
		}
		return err
	}
	return response.Created(c, toUserResponse(user))
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, &response.ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request body",
		})
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, &response.ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Validation failed",
			Details: sharevalidator.FieldErrors(err),
		})
	}
	access, refresh, user, err := h.usecase.Login(c.UserContext(), req, c.Get(fiber.HeaderUserAgent), c.IP())
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidCredentials) {
			return response.Error(c, fiber.StatusUnauthorized, &response.ErrorBody{
				Code:    "INVALID_CREDENTIALS",
				Message: apperrors.ErrInvalidCredentials.Message,
			})
		}
		return err
	}
	writeRefreshCookie(c, h.cfg, refresh)
	return response.OK(c, dto.TokenResponse{
		AccessToken: access,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.cfg.JWT.AccessTTL.Seconds()),
		User:        toUserResponse(user),
	})
}

// Refresh handles POST /api/v1/auth/refresh (reads refresh cookie).
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	refresh := c.Cookies(h.cfg.Cookie.Name)
	if refresh == "" {
		return response.Error(c, fiber.StatusUnauthorized, &response.ErrorBody{
			Code:    "MISSING_REFRESH_TOKEN",
			Message: "Refresh token required",
		})
	}
	access, newRefresh, err := h.usecase.Refresh(c.UserContext(), refresh, c.Get(fiber.HeaderUserAgent), c.IP())
	if err != nil {
		if errors.Is(err, apperrors.ErrInvalidRefreshToken) {
			clearRefreshCookie(c, h.cfg)
			return response.Error(c, fiber.StatusUnauthorized, &response.ErrorBody{
				Code:    "INVALID_REFRESH_TOKEN",
				Message: apperrors.ErrInvalidRefreshToken.Message,
			})
		}
		if errors.Is(err, apperrors.ErrSessionRevoked) {
			clearRefreshCookie(c, h.cfg)
			return response.Error(c, fiber.StatusUnauthorized, &response.ErrorBody{
				Code:    "SESSION_REVOKED",
				Message: "Session revoked",
			})
		}
		return err
	}
	writeRefreshCookie(c, h.cfg, newRefresh)
	return response.OK(c, dto.RefreshResponse{
		AccessToken: access,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.cfg.JWT.AccessTTL.Seconds()),
	})
}

// Logout handles POST /api/v1/auth/logout (hybrid idempotent: Bearer preferred, cookie fallback).
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	bearerUserID, bearerSessionID, hasBearer := bearerSessionFromCtx(c)
	h.usecase.LogoutHybrid(c.UserContext(), bearerUserID, bearerSessionID, hasBearer, c.Cookies(h.cfg.Cookie.Name))
	clearRefreshCookie(c, h.cfg)
	return response.NoContent(c)
}

// Me handles GET /api/v1/auth/me.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, &response.ErrorBody{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
	}
	user, err := h.usecase.Me(c.UserContext(), userID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			return response.Error(c, fiber.StatusUnauthorized, &response.ErrorBody{
				Code:    "SESSION_REVOKED",
				Message: "Session revoked",
			})
		}
		return err
	}
	return response.OK(c, toUserResponse(user))
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
