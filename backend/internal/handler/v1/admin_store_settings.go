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

type AdminStoreSettingsHandler struct {
	usecase *usecase.AdminStoreSettingsUsecase
}

func NewAdminStoreSettingsHandler(uc *usecase.AdminStoreSettingsUsecase) *AdminStoreSettingsHandler {
	return &AdminStoreSettingsHandler{usecase: uc}
}

// Get handles GET /api/v1/admin/settings.
func (h *AdminStoreSettingsHandler) Get(c fiber.Ctx) error {
	out, err := h.usecase.GetSettings(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Patch handles PATCH /api/v1/admin/settings.
func (h *AdminStoreSettingsHandler) Patch(c fiber.Ctx) error {
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	var req dto.PatchStoreSettingsRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	out, err := h.usecase.PatchSettings(c.Context(), actorID, actorRole, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// ListClosedDates handles GET /api/v1/admin/closed-dates.
func (h *AdminStoreSettingsHandler) ListClosedDates(c fiber.Ctx) error {
	out, err := h.usecase.ListClosedDates(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// CreateClosedDate handles POST /api/v1/admin/closed-dates.
func (h *AdminStoreSettingsHandler) CreateClosedDate(c fiber.Ctx) error {
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	var req dto.CreateClosedDateRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.AddClosedDate(c.Context(), actorID, actorRole, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, out)
}

// DeleteClosedDate handles DELETE /api/v1/admin/closed-dates/:id.
func (h *AdminStoreSettingsHandler) DeleteClosedDate(c fiber.Ctx) error {
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "must be a UUID"))
	}
	if err := h.usecase.RemoveClosedDate(c.Context(), actorID, actorRole, id); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}
