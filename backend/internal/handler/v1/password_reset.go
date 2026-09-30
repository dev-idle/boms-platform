package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
)

// PasswordResetHandler serves asking for a reset link and choosing a new
// password with it.
type PasswordResetHandler struct {
	usecase *usecase.PasswordResetUsecase
}

func NewPasswordResetHandler(uc *usecase.PasswordResetUsecase) *PasswordResetHandler {
	return &PasswordResetHandler{usecase: uc}
}

// Request handles POST /api/v1/auth/password-reset/request. It answers 202
// whether or not the address has an account.
func (h *PasswordResetHandler) Request(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	noStore(c)
	var req dto.PasswordResetRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	if err := h.usecase.Request(c.Context(), req.Email); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Accepted(c)
}

// Confirm handles POST /api/v1/auth/password-reset/confirm.
func (h *PasswordResetHandler) Confirm(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	noStore(c)
	var req dto.PasswordResetConfirmRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	if err := h.usecase.Confirm(c.Context(), req.Token, req.NewPassword); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}
