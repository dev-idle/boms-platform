package v1

import (
	"errors"

	domainuser "github.com/boms/backend/internal/domain/user"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// writeMapUsecaseError maps known domain and AppError values to HTTP responses.
// Unmapped errors propagate to fiber.ErrorHandler as internal_error with logging.
func writeMapUsecaseError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, usecase.ErrEmailExists):
		return writeAppError(c, usecase.ErrEmailExists)
	case errors.Is(err, usecase.ErrMeNotFound):
		return writeAppError(c, usecase.ErrMeNotFound)
	case errors.Is(err, usecase.ErrUserNotFound):
		return writeAppError(c, usecase.ErrUserNotFound)
	case errors.Is(err, apperrors.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, apperrors.ErrConflict):
		return writeAppError(c, apperrors.ErrConflict)
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		return writeAppError(c, apperrors.ErrInvalidCredentials)
	case errors.Is(err, apperrors.ErrInvalidRefreshToken):
		return writeAppError(c, apperrors.ErrInvalidRefreshToken)
	case errors.Is(err, apperrors.ErrMissingRefreshToken):
		return writeAppError(c, apperrors.ErrMissingRefreshToken)
	case errors.Is(err, apperrors.ErrSessionRevoked):
		return writeAppError(c, apperrors.ErrSessionRevoked)
	case errors.Is(err, apperrors.ErrTokenExpired):
		return writeAppError(c, apperrors.ErrTokenExpired)
	case errors.Is(err, apperrors.ErrForbidden):
		return writeAppError(c, apperrors.ErrForbidden)
	case errors.Is(err, domainuser.ErrProfileNotFound):
		return writeAppError(c, apperrors.ErrProfileNotFound)
	case errors.Is(err, domainuser.ErrEmployeeCodeExists):
		return writeAppError(c, apperrors.ErrEmployeeCodeExists)
	case errors.Is(err, domainuser.ErrCannotModifySelf):
		return writeAppError(c, apperrors.ErrCannotModifySelf)
	case errors.Is(err, domainuser.ErrInvalidRoleTransition):
		return writeAppError(c, apperrors.ErrInvalidRoleTransition)
	}
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return writeAppError(c, appErr)
	}
	return err
}

// mustChangePasswordPtr returns a JSON pointer when the flag is true (omitempty otherwise).
func mustChangePasswordPtr(v bool) *bool {
	if !v {
		return nil
	}
	t := true
	return &t
}
