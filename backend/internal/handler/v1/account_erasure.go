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

// AccountErasureHandler serves DELETE /me: a customer erasing their account.
type AccountErasureHandler struct {
	usecase *usecase.AccountErasureUsecase
}

func NewAccountErasureHandler(uc *usecase.AccountErasureUsecase) *AccountErasureHandler {
	return &AccountErasureHandler{usecase: uc}
}

func (h *AccountErasureHandler) Erase(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	var req dto.EraseMyAccountRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	if err := h.usecase.Erase(c.Context(), userID, req.Password); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}
