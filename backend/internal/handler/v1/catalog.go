package v1

import (
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CatalogHandler struct {
	usecase *usecase.CatalogUsecase
}

func NewCatalogHandler(uc *usecase.CatalogUsecase) *CatalogHandler {
	return &CatalogHandler{usecase: uc}
}

func (h *CatalogHandler) ListCategories(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.CatalogListDefaultPageSizeQuery),
		usecase.CatalogListDefaultPageSize,
	)

	items, total, page, pageSize, err := h.usecase.ListCategories(c.UserContext(), page, pageSize)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *CatalogHandler) ListProducts(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.CatalogListDefaultPageSizeQuery),
		usecase.CatalogListDefaultPageSize,
	)
	categoryID := c.Query("category_id", "")

	items, total, page, pageSize, err := h.usecase.ListProducts(c.UserContext(), page, pageSize, categoryID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *CatalogHandler) GetProduct(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid product id"))
	}
	out, err := h.usecase.GetProduct(c.UserContext(), id)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *CatalogHandler) ListCombos(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.CatalogListDefaultPageSizeQuery),
		usecase.CatalogListDefaultPageSize,
	)

	items, total, page, pageSize, err := h.usecase.ListCombos(c.UserContext(), page, pageSize)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *CatalogHandler) GetCombo(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid combo id"))
	}
	out, err := h.usecase.GetCombo(c.UserContext(), id)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
