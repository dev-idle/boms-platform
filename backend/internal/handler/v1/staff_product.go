package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
)

type StaffProductHandler struct {
	usecase *usecase.StaffProductUsecase
}

func NewStaffProductHandler(uc *usecase.StaffProductUsecase) *StaffProductHandler {
	return &StaffProductHandler{usecase: uc}
}

// List pages the products the counter sells; ?sold_out_today=true keeps those
// out today.
func (h *StaffProductHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.CatalogListDefaultPageSizeQuery),
		usecase.CatalogListDefaultPageSize,
	)
	out, total, page, pageSize, err := h.usecase.List(c.Context(), page, pageSize, c.Query("sold_out_today") == "true")
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, out, int(page), int(pageSize), total)
}

// PatchSoldOut marks a product sold out for today, or back.
func (h *StaffProductHandler) PatchSoldOut(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid product id"))
	}
	var req dto.PatchStaffProductSoldOutRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.SetSoldOut(c.Context(), actorID, actorRole, productID, *req.SoldOut)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
