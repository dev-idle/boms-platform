package v1

import (
	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ManagerDiscountCodeHandler struct {
	usecase *usecase.ManagerDiscountCodeUsecase
}

func NewManagerDiscountCodeHandler(uc *usecase.ManagerDiscountCodeUsecase) *ManagerDiscountCodeHandler {
	return &ManagerDiscountCodeHandler{usecase: uc}
}

func (h *ManagerDiscountCodeHandler) List(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.CatalogListDefaultPageSizeQuery),
		usecase.CatalogListDefaultPageSize,
	)
	search := c.Query("search", "")

	items, total, page, pageSize, err := h.usecase.List(c.UserContext(), page, pageSize, search)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *ManagerDiscountCodeHandler) Get(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid discount code id"))
	}
	out, err := h.usecase.Get(c.UserContext(), id)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *ManagerDiscountCodeHandler) Create(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	var req dto.CreateDiscountCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Create(c.UserContext(), actorID, actorRole, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, out)
}

func (h *ManagerDiscountCodeHandler) Patch(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid discount code id"))
	}
	var req dto.UpdateDiscountCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Update(c.UserContext(), actorID, actorRole, id, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *ManagerDiscountCodeHandler) Delete(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid discount code id"))
	}
	if err := h.usecase.Delete(c.UserContext(), actorID, actorRole, id); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}
