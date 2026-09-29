package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
)

// BakerTicketHandler serves the kitchen's ticket queue.
type BakerTicketHandler struct {
	usecase *usecase.BakerTicketUsecase
}

func NewBakerTicketHandler(uc *usecase.BakerTicketUsecase) *BakerTicketHandler {
	return &BakerTicketHandler{usecase: uc}
}

func (h *BakerTicketHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	page, pageSize := ticketPage(c)
	items, total, page, pageSize, err := h.usecase.List(c.Context(), page, pageSize, c.Query("status", ""))
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *BakerTicketHandler) Get(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid ticket id"))
	}
	out, err := h.usecase.Get(c.Context(), ticketID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *BakerTicketHandler) PatchStatus(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid ticket id"))
	}
	var req dto.PatchTicketStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.PatchStatus(c.Context(), actorID, actorRole, ticketID, domainorder.TicketStatus(req.Status))
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// StaffTicketHandler serves the counter's prep queue and ticket moves.
type StaffTicketHandler struct {
	usecase *usecase.StaffTicketUsecase
}

func NewStaffTicketHandler(uc *usecase.StaffTicketUsecase) *StaffTicketHandler {
	return &StaffTicketHandler{usecase: uc}
}

func (h *StaffTicketHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	page, pageSize := ticketPage(c)
	items, total, page, pageSize, err := h.usecase.List(c.Context(), page, pageSize, c.Query("status", ""))
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

func (h *StaffTicketHandler) PatchStatus(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid ticket id"))
	}
	var req dto.PatchTicketStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.PatchStatus(c.Context(), actorID, actorRole, ticketID, domainorder.TicketStatus(req.Status))
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *StaffTicketHandler) Move(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	ticketID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid ticket id"))
	}
	var req dto.MoveTicketRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Move(c.Context(), actorID, actorRole, ticketID, domaincategory.Station(req.Station))
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func ticketPage(c fiber.Ctx) (int32, int32) {
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.OrderListDefaultPageSizeQuery),
		usecase.OrderListDefaultPageSize,
	)
	return page, pageSize
}
