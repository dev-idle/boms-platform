package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
)

// ManagerOrderHandler serves the manager's handover of an online order whose
// customer cannot give its pickup code.
type ManagerOrderHandler struct {
	usecase *usecase.ManagerOrderUsecase
}

func NewManagerOrderHandler(uc *usecase.ManagerOrderUsecase) *ManagerOrderHandler {
	return &ManagerOrderHandler{usecase: uc}
}

// Lookup handles GET /api/v1/manager/orders/lookup?code=.
func (h *ManagerOrderHandler) Lookup(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	code := c.Query("code")
	if code == "" {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("code", "required"))
	}
	out, err := h.usecase.Lookup(c.Context(), code)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// HandOver handles POST /api/v1/manager/orders/:id/handover.
func (h *ManagerOrderHandler) HandOver(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	var req dto.ManagerHandoverRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.HandOver(c.Context(), actorID, actorRole, orderID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
