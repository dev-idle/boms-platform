package v1

import (
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/gofiber/fiber/v2"
)

// writeAppError converts an AppError to the canonical response.Error envelope.
func writeAppError(c *fiber.Ctx, e *apperrors.AppError) error {
	code, message, details := e.ToErrorBody()
	return response.Error(c, e.StatusCode, &response.ErrorBody{
		Code:    code,
		Message: message,
		Details: details,
	})
}

// writeValidationError maps validator failures to 400 validation_error with field details.
func writeValidationError(c *fiber.Ctx, err error) error {
	return response.Error(c, apperrors.ErrValidation.StatusCode, &response.ErrorBody{
		Code:    apperrors.ErrValidation.Code,
		Message: apperrors.ErrValidation.Message,
		Details: sharevalidator.FieldErrors(err),
	})
}
