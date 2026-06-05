package v1

import (
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/middleware"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type OrderHandler struct {
	usecase *usecase.OrderUsecase
}

func NewOrderHandler(uc *usecase.OrderUsecase) *OrderHandler {
	return &OrderHandler{usecase: uc}
}

func (h *OrderHandler) Checkout(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	out, err := h.usecase.Checkout(c.UserContext(), userID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *OrderHandler) List(c *fiber.Ctx) error {
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
	items, total, page, pageSize, err := h.usecase.List(c.UserContext(), userID, page, pageSize)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *OrderHandler) Get(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	out, err := h.usecase.Get(c.UserContext(), userID, orderID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
