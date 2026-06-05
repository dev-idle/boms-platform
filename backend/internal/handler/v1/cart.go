package v1

import (
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CartHandler struct {
	usecase *usecase.CartUsecase
}

func NewCartHandler(uc *usecase.CartUsecase) *CartHandler {
	return &CartHandler{usecase: uc}
}

func (h *CartHandler) Get(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	out, err := h.usecase.Get(c.UserContext(), userID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *CartHandler) AddItem(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	var req dto.AddCartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.AddItem(c.UserContext(), userID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *CartHandler) UpdateItem(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	itemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid cart item id"))
	}
	var req dto.UpdateCartItemRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.UpdateItem(c.UserContext(), userID, itemID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *CartHandler) RemoveItem(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	itemID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid cart item id"))
	}
	out, err := h.usecase.RemoveItem(c.UserContext(), userID, itemID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *CartHandler) ApplyDiscount(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	var req dto.ApplyCartDiscountRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.ApplyDiscount(c.UserContext(), userID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *CartHandler) RemoveDiscount(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	out, err := h.usecase.RemoveDiscount(c.UserContext(), userID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
