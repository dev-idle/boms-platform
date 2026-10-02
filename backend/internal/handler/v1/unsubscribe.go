package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
)

// UnsubscribeHandler is the unsubscribe link of a promotion email.
type UnsubscribeHandler struct {
	usecase *usecase.UnsubscribeUsecase
}

func NewUnsubscribeHandler(uc *usecase.UnsubscribeUsecase) *UnsubscribeHandler {
	return &UnsubscribeHandler{usecase: uc}
}

// Unsubscribe stops promotion emails to the customer the link's token names.
func (h *UnsubscribeHandler) Unsubscribe(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	var req dto.UnsubscribeRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	if err := h.usecase.Unsubscribe(c.Context(), req.Token); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}
