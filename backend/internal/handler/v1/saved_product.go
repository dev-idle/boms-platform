package v1

import (
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// SavedProductHandler is a customer's favorites and wishlist.
type SavedProductHandler struct {
	usecase *usecase.SavedProductUsecase
}

func NewSavedProductHandler(uc *usecase.SavedProductUsecase) *SavedProductHandler {
	return &SavedProductHandler{usecase: uc}
}

// List returns both of the customer's lists.
func (h *SavedProductHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	out, err := h.usecase.List(c.Context(), userID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Save puts a product on a list (/:list/:product_id).
func (h *SavedProductHandler) Save(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("product_id", "invalid product id"))
	}
	if err := h.usecase.Save(c.Context(), userID, c.Params("list"), productID); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}

// Remove takes a product off a list (/:list/:product_id).
func (h *SavedProductHandler) Remove(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	productID, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("product_id", "invalid product id"))
	}
	if err := h.usecase.Remove(c.Context(), userID, c.Params("list"), productID); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}
