package v1

import (
	"errors"
	"strconv"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
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
		return h.mapError(c, err)
	}
	return response.OK(c, out)
}

func (h *AdminUserHandler) List(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))
	search := c.Query("search", "")

	items, total, err := h.usecase.List(c.UserContext(), page, pageSize, search)
	if err != nil {
		return err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	return response.OKPaginated(c, items, page, pageSize, total)
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
		return h.mapError(c, err)
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
		return h.mapError(c, err)
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
		return h.mapError(c, err)
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
		return h.mapError(c, err)
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
		return h.mapError(c, err)
	}
	return response.NoContent(c)
}

func (h *AdminUserHandler) mapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domainuser.ErrCannotModifySelf):
		return writeAppError(c, apperrors.ErrCannotModifySelf)
	case errors.Is(err, domainuser.ErrInvalidRoleTransition):
		return writeAppError(c, apperrors.ErrInvalidRoleTransition)
	case errors.Is(err, domainuser.ErrEmployeeCodeExists):
		return writeAppError(c, apperrors.ErrEmployeeCodeExists)
	case errors.Is(err, apperrors.ErrNotFound):
		return writeAppError(c, apperrors.ErrNotFound)
	case errors.Is(err, apperrors.ErrConflict):
		return writeAppError(c, apperrors.ErrConflict)
	default:
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return writeAppError(c, appErr)
		}
		return err
	}
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
