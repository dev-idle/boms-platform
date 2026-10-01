package v1

import (
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// ConversationHandler is a customer's messages about their orders.
type ConversationHandler struct {
	usecase *usecase.ConversationUsecase
}

func NewConversationHandler(uc *usecase.ConversationUsecase) *ConversationHandler {
	return &ConversationHandler{usecase: uc}
}

// Thread returns a page of the messages on the customer's order
// (?before=<message id> for older ones).
func (h *ConversationHandler) Thread(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	beforeID, appErr := parseBeforeMessage(c)
	if appErr != nil {
		return writeAppError(c, appErr)
	}
	out, err := h.usecase.Thread(c.Context(), userID, orderID, beforeID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Post writes the customer's message on their order.
func (h *ConversationHandler) Post(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
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
	out, err := h.usecase.Post(c.Context(), userID, orderID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, out)
}

// MarkRead records that the customer read the messages on their order.
func (h *ConversationHandler) MarkRead(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	if err := h.usecase.MarkRead(c.Context(), userID, orderID); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}

// parseBeforeMessage reads the optional ?before= message id a thread page
// starts before.
func parseBeforeMessage(c fiber.Ctx) (*uuid.UUID, *apperrors.AppError) {
	raw := c.Query("before")
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("before", "invalid message id")
	}
	return &id, nil
}
