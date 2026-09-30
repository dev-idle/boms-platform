package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
)

// EmailVerificationHandler serves confirming an address from its emailed link
// and asking for a new link.
type EmailVerificationHandler struct {
	usecase *usecase.EmailVerificationUsecase
}

func NewEmailVerificationHandler(uc *usecase.EmailVerificationUsecase) *EmailVerificationHandler {
	return &EmailVerificationHandler{usecase: uc}
}

// Verify handles POST /api/v1/auth/verify-email.
func (h *EmailVerificationHandler) Verify(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	noStore(c)
	var req dto.VerifyEmailRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	if err := h.usecase.Verify(c.Context(), req.Token); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}

// Resend handles POST /api/v1/me/email-verification.
func (h *EmailVerificationHandler) Resend(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	if err := h.usecase.Resend(c.Context(), userID); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Accepted(c)
}
