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

// ManagerReviewHandler is a manager moderating reviews and reading feedback.
type ManagerReviewHandler struct {
	usecase *usecase.ManagerReviewUsecase
}

func NewManagerReviewHandler(uc *usecase.ManagerReviewUsecase) *ManagerReviewHandler {
	return &ManagerReviewHandler{usecase: uc}
}

// List pages the reviews, latest first; ?status= narrows them.
func (h *ManagerReviewHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.CatalogListDefaultPageSizeQuery),
		usecase.CatalogListDefaultPageSize,
	)
	out, total, page, pageSize, err := h.usecase.List(c.Context(), page, pageSize, c.Query("status"))
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, out, int(page), int(pageSize), total)
}

// Summary returns the customers' feedback added up.
func (h *ManagerReviewHandler) Summary(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.Summary(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Moderate publishes a review or hides it.
func (h *ManagerReviewHandler) Moderate(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	reviewID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid review id"))
	}
	var req dto.ModerateReviewRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Moderate(c.Context(), actorID, actorRole, reviewID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
