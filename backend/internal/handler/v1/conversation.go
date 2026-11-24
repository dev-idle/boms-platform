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

// CustomerConversationParam reads which of the customer's conversations a
// route names: the order's, or the general one (nil).
type CustomerConversationParam func(c fiber.Ctx) (*uuid.UUID, *apperrors.AppError)

// OrderConversation is the conversation about the order the route's :id names.
func OrderConversation(c fiber.Ctx) (*uuid.UUID, *apperrors.AppError) {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("id", "invalid order id")
	}
	return &orderID, nil
}

// GeneralConversation is the customer's general conversation.
func GeneralConversation(fiber.Ctx) (*uuid.UUID, *apperrors.AppError) {
	return nil, nil
}

// ConversationHandler is a customer's messages with the counter.
type ConversationHandler struct {
	usecase *usecase.ConversationUsecase
}

func NewConversationHandler(uc *usecase.ConversationUsecase) *ConversationHandler {
	return &ConversationHandler{usecase: uc}
}

// Thread returns a page of the conversation's messages (?before=<message id>
// for older ones).
func (h *ConversationHandler) Thread(param CustomerConversationParam) fiber.Handler {
	return func(c fiber.Ctx) error {
		response.EnsureRequestID(c)
		userID, orderID, appErr := customerConversation(c, param)
		if appErr != nil {
			return writeAppError(c, appErr)
		}
		beforeID, appErr := parseBeforeID(c, "message")
		if appErr != nil {
			return writeAppError(c, appErr)
		}
		out, err := h.usecase.Thread(c.Context(), userID, orderID, beforeID)
		if err != nil {
			return writeMapUsecaseError(c, err)
		}
		return response.OK(c, out)
	}
}

// Post writes the customer's message on the conversation.
func (h *ConversationHandler) Post(param CustomerConversationParam) fiber.Handler {
	return func(c fiber.Ctx) error {
		response.EnsureRequestID(c)
		userID, orderID, appErr := customerConversation(c, param)
		if appErr != nil {
			return writeAppError(c, appErr)
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
}

// MarkRead records that the customer read the conversation's messages.
func (h *ConversationHandler) MarkRead(param CustomerConversationParam) fiber.Handler {
	return func(c fiber.Ctx) error {
		response.EnsureRequestID(c)
		userID, orderID, appErr := customerConversation(c, param)
		if appErr != nil {
			return writeAppError(c, appErr)
		}
		if err := h.usecase.MarkRead(c.Context(), userID, orderID); err != nil {
			return writeMapUsecaseError(c, err)
		}
		return response.NoContent(c)
	}
}

// customerConversation reads the signed-in customer and the conversation
// param names.
func customerConversation(c fiber.Ctx, param CustomerConversationParam) (uuid.UUID, *uuid.UUID, *apperrors.AppError) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return uuid.Nil, nil, apperrors.ErrUnauthorized
	}
	orderID, appErr := param(c)
	return userID, orderID, appErr
}

// parseBeforeID reads the optional ?before= id of the message or review a
// page starts before.
func parseBeforeID(c fiber.Ctx, noun string) (*uuid.UUID, *apperrors.AppError) {
	raw := c.Query("before")
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("before", "invalid "+noun+" id")
	}
	return &id, nil
}
