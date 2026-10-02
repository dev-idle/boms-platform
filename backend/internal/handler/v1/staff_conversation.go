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

// StaffConversationHandler is the counter's inbox and its replies.
type StaffConversationHandler struct {
	usecase *usecase.StaffConversationUsecase
}

func NewStaffConversationHandler(uc *usecase.StaffConversationUsecase) *StaffConversationHandler {
	return &StaffConversationHandler{usecase: uc}
}

// List pages the inbox (?status=open|closed).
func (h *StaffConversationHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.OrderListDefaultPageSizeQuery),
		usecase.OrderListDefaultPageSize,
	)
	items, total, page, pageSize, err := h.usecase.List(c.Context(), page, pageSize, c.Query("status"))
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

// Counts returns how many conversations are open and unread.
func (h *StaffConversationHandler) Counts(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.Counts(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Thread returns a page of the messages on an order (?before=<message id>).
func (h *StaffConversationHandler) Thread(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	beforeID, appErr := parseBeforeID(c, "message")
	if appErr != nil {
		return writeAppError(c, appErr)
	}
	out, err := h.usecase.Thread(c.Context(), orderID, beforeID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Post writes a staff member's message on an order.
func (h *StaffConversationHandler) Post(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	var req dto.PostMessageRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Post(c.Context(), actorID, orderID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, out)
}

// MarkRead records that the counter read the messages on an order.
func (h *StaffConversationHandler) MarkRead(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	if err := h.usecase.MarkRead(c.Context(), orderID); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}

// PatchStatus resolves an order's conversation, or opens it again.
func (h *StaffConversationHandler) PatchStatus(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	var req dto.PatchConversationRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.SetStatus(c.Context(), orderID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
