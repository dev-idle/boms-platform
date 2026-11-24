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

// StaffConversationParam reads which conversation a counter route names.
type StaffConversationParam func(c fiber.Ctx) (usecase.StaffThreadKey, *apperrors.AppError)

// StaffOrderConversation is the conversation about the order the route's :id
// names.
func StaffOrderConversation(c fiber.Ctx) (usecase.StaffThreadKey, *apperrors.AppError) {
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return usecase.StaffThreadKey{}, apperrors.ErrValidation.WithDetail("id", "invalid order id")
	}
	return usecase.OrderThread(orderID), nil
}

// StaffGeneralConversation is the general conversation of the customer the
// route's :id names.
func StaffGeneralConversation(c fiber.Ctx) (usecase.StaffThreadKey, *apperrors.AppError) {
	customerID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return usecase.StaffThreadKey{}, apperrors.ErrValidation.WithDetail("id", "invalid customer id")
	}
	return usecase.GeneralThread(customerID), nil
}

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

// Thread returns a page of a conversation's messages (?before=<message id>).
func (h *StaffConversationHandler) Thread(param StaffConversationParam) fiber.Handler {
	return func(c fiber.Ctx) error {
		response.EnsureRequestID(c)
		key, appErr := param(c)
		if appErr != nil {
			return writeAppError(c, appErr)
		}
		beforeID, appErr := parseBeforeID(c, "message")
		if appErr != nil {
			return writeAppError(c, appErr)
		}
		out, err := h.usecase.Thread(c.Context(), key, beforeID)
		if err != nil {
			return writeMapUsecaseError(c, err)
		}
		return response.OK(c, out)
	}
}

// Post writes a staff member's message on a conversation.
func (h *StaffConversationHandler) Post(param StaffConversationParam) fiber.Handler {
	return func(c fiber.Ctx) error {
		response.EnsureRequestID(c)
		actorID, ok := middleware.GetUserID(c)
		if !ok {
			return writeAppError(c, apperrors.ErrUnauthorized)
		}
		key, appErr := param(c)
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
		out, err := h.usecase.Post(c.Context(), actorID, key, req)
		if err != nil {
			return writeMapUsecaseError(c, err)
		}
		return response.Created(c, out)
	}
}

// MarkRead records that the counter read a conversation's messages.
func (h *StaffConversationHandler) MarkRead(param StaffConversationParam) fiber.Handler {
	return func(c fiber.Ctx) error {
		response.EnsureRequestID(c)
		key, appErr := param(c)
		if appErr != nil {
			return writeAppError(c, appErr)
		}
		if err := h.usecase.MarkRead(c.Context(), key); err != nil {
			return writeMapUsecaseError(c, err)
		}
		return response.NoContent(c)
	}
}

// PatchStatus resolves a conversation, or opens it again.
func (h *StaffConversationHandler) PatchStatus(param StaffConversationParam) fiber.Handler {
	return func(c fiber.Ctx) error {
		response.EnsureRequestID(c)
		key, appErr := param(c)
		if appErr != nil {
			return writeAppError(c, appErr)
		}
		var req dto.PatchConversationRequest
		if err := c.Bind().Body(&req); err != nil {
			return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
		}
		if err := sharevalidator.Struct(&req); err != nil {
			return writeValidationError(c, err)
		}
		out, err := h.usecase.SetStatus(c.Context(), key, req)
		if err != nil {
			return writeMapUsecaseError(c, err)
		}
		return response.OK(c, out)
	}
}
