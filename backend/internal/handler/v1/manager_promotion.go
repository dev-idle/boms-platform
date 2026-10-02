package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
)

// ManagerPromotionHandler is a manager emailing promotions.
type ManagerPromotionHandler struct {
	usecase *usecase.ManagerPromotionUsecase
}

func NewManagerPromotionHandler(uc *usecase.ManagerPromotionUsecase) *ManagerPromotionHandler {
	return &ManagerPromotionHandler{usecase: uc}
}

// List pages the promotions sent, latest first.
func (h *ManagerPromotionHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.CatalogListDefaultPageSizeQuery),
		usecase.CatalogListDefaultPageSize,
	)
	out, total, page, pageSize, err := h.usecase.List(c.Context(), page, pageSize)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, out, int(page), int(pageSize), total)
}

// Audience returns how many customers a promotion sent now would go to.
func (h *ManagerPromotionHandler) Audience(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.Audience(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Send records a promotion for the worker to email.
func (h *ManagerPromotionHandler) Send(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	var req dto.SendPromotionRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Send(c.Context(), actorID, actorRole, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, out)
}
