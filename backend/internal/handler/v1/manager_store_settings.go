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

type ManagerStoreSettingsHandler struct {
	usecase *usecase.ManagerStoreSettingsUsecase
}

func NewManagerStoreSettingsHandler(uc *usecase.ManagerStoreSettingsUsecase) *ManagerStoreSettingsHandler {
	return &ManagerStoreSettingsHandler{usecase: uc}
}

// Get handles GET /api/v1/manager/settings.
func (h *ManagerStoreSettingsHandler) Get(c fiber.Ctx) error {
	out, err := h.usecase.GetSettings(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Patch handles PATCH /api/v1/manager/settings.
func (h *ManagerStoreSettingsHandler) Patch(c fiber.Ctx) error {
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

// ListClosedDates handles GET /api/v1/manager/closed-dates.
func (h *ManagerStoreSettingsHandler) ListClosedDates(c fiber.Ctx) error {
	out, err := h.usecase.ListClosedDates(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// CreateClosedDate handles POST /api/v1/manager/closed-dates.
func (h *ManagerStoreSettingsHandler) CreateClosedDate(c fiber.Ctx) error {
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

// StrandedPickups handles GET /api/v1/manager/settings/stranded-pickups: the
// orders a closed day (?closed_date=) or new hours (?opens_at=&closes_at=)
// would strand.
func (h *ManagerStoreSettingsHandler) StrandedPickups(c fiber.Ctx) error {
	out, err := h.usecase.StrandedPickups(c.Context(), dto.StrandedPickupsQuery{
		ClosedDate: c.Query("closed_date"), OpensAt: c.Query("opens_at"), ClosesAt: c.Query("closes_at"),
	})
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// CancelStrandedPickups handles POST /api/v1/manager/settings/stranded-pickups/cancel.
func (h *ManagerStoreSettingsHandler) CancelStrandedPickups(c fiber.Ctx) error {
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	var req dto.CancelStrandedPickupsRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.CancelStrandedPickups(c.Context(), actorID, actorRole, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// DeleteClosedDate handles DELETE /api/v1/manager/closed-dates/:id.
func (h *ManagerStoreSettingsHandler) DeleteClosedDate(c fiber.Ctx) error {
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
