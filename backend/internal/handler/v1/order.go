package v1

import (
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type OrderHandler struct {
	usecase *usecase.OrderUsecase
}

func NewOrderHandler(uc *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{usecase: uc}
}

func (h *OrderHandler) Checkout(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	// Every checkout carries one: a retry with the same key returns the order
	// the first attempt placed.
	checkoutKey, err := uuid.Parse(c.Get("Idempotency-Key"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("idempotency_key", "must be a UUID"))
	}
	var req dto.CheckoutRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Checkout(c.Context(), userID, checkoutKey, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *OrderHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.OrderListDefaultPageSizeQuery),
		usecase.OrderListDefaultPageSize,
	)
	items, total, page, pageSize, err := h.usecase.List(c.Context(), userID, page, pageSize, dto.OrderHistoryQuery{
		Status: c.Query("status"),
		From:   c.Query("from"),
		To:     c.Query("to"),
	})
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *OrderHandler) Get(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	out, err := h.usecase.Get(c.Context(), userID, orderID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Cancel cancels the customer's order while it is not being made yet.
func (h *OrderHandler) Cancel(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	out, err := h.usecase.Cancel(c.Context(), userID, orderID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Reschedule moves the pickup of the customer's order while it is not being
// made yet.
func (h *OrderHandler) Reschedule(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	var req dto.RescheduleOrderRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Reschedule(c.Context(), userID, orderID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
