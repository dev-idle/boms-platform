package v1

import (
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AdminUserHandler struct {
	usecase *usecase.AdminUserUsecase
}

func NewAdminUserHandler(uc *usecase.AdminUserUsecase) *AdminUserHandler {
	return &AdminUserHandler{usecase: uc}
}

func (h *AdminUserHandler) Get(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid user id"))
	}
	out, err := h.usecase.Get(c.UserContext(), targetID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *AdminUserHandler) ListActivity(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid user id"))
	}
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.AdminUserActivityDefaultPageSizeQuery),
		usecase.AdminUserActivityDefaultPageSize,
	)
	items, total, page, pageSize, err := h.usecase.ListActivity(c.UserContext(), targetID, page, pageSize)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *AdminUserHandler) List(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.AdminUserListDefaultPageSizeQuery),
		usecase.AdminUserListDefaultPageSize,
	)
	search := c.Query("search", "")
	role := c.Query("role", "")

	items, total, page, pageSize, err := h.usecase.List(c.UserContext(), page, pageSize, search, role)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *AdminUserHandler) Create(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	var req dto.CreateOperationalUserRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.CreateOperationalUser(c.UserContext(), actorID, actorRole, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, out)
}

func (h *AdminUserHandler) PatchProfile(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid user id"))
	}
	var req dto.UpdateOperationalProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.UpdateOperationalProfile(c.UserContext(), actorID, actorRole, targetID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *AdminUserHandler) PatchRole(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid user id"))
	}
	var req dto.UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.UpdateRole(c.UserContext(), actorID, actorRole, targetID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *AdminUserHandler) PatchEnable(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid user id"))
	}
	if err := h.usecase.Enable(c.UserContext(), actorID, actorRole, targetID); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}

func (h *AdminUserHandler) PostResetPassword(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid user id"))
	}
	out, err := h.usecase.ResetPassword(c.UserContext(), actorID, actorRole, targetID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *AdminUserHandler) PatchDisable(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid user id"))
	}
	if err := h.usecase.Disable(c.UserContext(), actorID, actorRole, targetID); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}

func (h *AdminUserHandler) RevokeSessions(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	targetID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid user id"))
	}
	if err := h.usecase.RevokeSessions(c.UserContext(), actorID, actorRole, targetID); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}

func actorFromCtx(c *fiber.Ctx) (uuid.UUID, domainuser.Role, error) {
	uid, ok := middleware.GetUserID(c)
	if !ok {
		return uuid.Nil, "", writeAppError(c, apperrors.ErrUnauthorized)
	}
	role, ok := middleware.GetRole(c)
	if !ok {
		return uuid.Nil, "", writeAppError(c, apperrors.ErrUnauthorized)
	}
	return uid, role, nil
}
