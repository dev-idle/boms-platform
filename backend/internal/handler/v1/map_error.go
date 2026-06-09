package v1

import (
	"errors"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domaincart "github.com/boms/backend/internal/domain/cart"
	domaincombo "github.com/boms/backend/internal/domain/combo"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainproduct "github.com/boms/backend/internal/domain/product"
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
	case errors.Is(err, domainuser.ErrSelfDeleteCustomerOnly):
		return writeAppError(c, apperrors.ErrSelfDeleteCustomerOnly)
	case errors.Is(err, domainuser.ErrProfileNotFound):
		return writeAppError(c, apperrors.ErrProfileNotFound)
	case errors.Is(err, domainuser.ErrEmployeeCodeExists):
		return writeAppError(c, apperrors.ErrEmployeeCodeExists)
	case errors.Is(err, domainuser.ErrCannotModifySelf):
		return writeAppError(c, apperrors.ErrCannotModifySelf)
	case errors.Is(err, domainuser.ErrInvalidRoleTransition):
		return writeAppError(c, apperrors.ErrInvalidRoleTransition)
	case errors.Is(err, domaincategory.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domaincategory.ErrHasProducts):
		return writeAppError(c, apperrors.ErrCategoryHasProducts)
	case errors.Is(err, domaincategory.ErrSlugExists):
		return writeAppError(c, apperrors.ErrSlugExists)
	case errors.Is(err, domaincategory.ErrInactive):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("category_id", "category is inactive"))
	case errors.Is(err, domainproduct.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domainproduct.ErrSlugExists):
		return writeAppError(c, apperrors.ErrSlugExists)
	case errors.Is(err, domaincombo.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domaincombo.ErrSlugExists):
		return writeAppError(c, apperrors.ErrSlugExists)
	case errors.Is(err, domaindiscount.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domaindiscount.ErrCodeExists):
		return writeAppError(c, apperrors.ErrCodeExists)
	case errors.Is(err, domaindiscount.ErrInactive):
		return writeAppError(c, apperrors.ErrDiscountInactive)
	case errors.Is(err, domaindiscount.ErrExpired):
		return writeAppError(c, apperrors.ErrDiscountExpired)
	case errors.Is(err, domaindiscount.ErrExhausted):
		return writeAppError(c, apperrors.ErrDiscountExhausted)
	case errors.Is(err, domaindiscount.ErrMinOrderNotMet):
		return writeAppError(c, apperrors.ErrDiscountMinOrderNotMet)
	case errors.Is(err, domaincart.ErrItemNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domaincart.ErrEmpty):
		return writeAppError(c, apperrors.ErrCartEmpty)
	case errors.Is(err, domaincart.ErrProductUnavailable):
		return writeAppError(c, apperrors.ErrProductUnavailable)
	case errors.Is(err, domaincart.ErrComboUnavailable):
		return writeAppError(c, apperrors.ErrComboUnavailable)
	case errors.Is(err, domaincart.ErrMaxItemsReached):
		return writeAppError(c, apperrors.ErrCartMaxItems)
	case errors.Is(err, domaincart.ErrQuantityOutOfRange):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("quantity", "must be between 1 and 99"))
	case errors.Is(err, domaincart.ErrInvalidLine):
		return writeAppError(c, apperrors.ErrValidation.WithDetail("line", "provide exactly one of product_id or combo_id"))
	case errors.Is(err, domainorder.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, domainorder.ErrInvalidStatusTransition):
		return writeAppError(c, apperrors.ErrInvalidOrderStatusTransition)
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
