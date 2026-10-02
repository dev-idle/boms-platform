package v1

import (
	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type StaffOrderHandler struct {
	usecase *usecase.StaffOrderUsecase
}

func NewStaffOrderHandler(uc *usecase.StaffOrderUsecase) *StaffOrderHandler {
	return &StaffOrderHandler{usecase: uc}
}

func (h *StaffOrderHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.OrderListDefaultPageSizeQuery),
		usecase.OrderListDefaultPageSize,
	)
	status := c.Query("status", "")

	items, total, page, pageSize, err := h.usecase.List(c.Context(), page, pageSize, status)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, items, int(page), int(pageSize), total)
}

// Pickups pages a bakery day's pickups (?date=YYYY-MM-DD) by time.
func (h *StaffOrderHandler) Pickups(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.OrderListDefaultPageSizeQuery),
		usecase.OrderListDefaultPageSize,
	)
	out, total, page, pageSize, err := h.usecase.Pickups(c.Context(), c.Query("date"), page, pageSize)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, out, int(page), int(pageSize), total)
}

func (h *StaffOrderHandler) Get(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	out, err := h.usecase.Get(c.Context(), orderID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

func (h *StaffOrderHandler) PatchStatus(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	var req dto.PatchStaffOrderStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.PatchStatus(c.Context(), actorID, actorRole, orderID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// ReportIncident records what staff found wrong with an order.
func (h *StaffOrderHandler) ReportIncident(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	var req dto.ReportOrderIncidentRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.ReportIncident(c.Context(), actorID, actorRole, orderID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, out)
}

// Create takes an order at the counter or on the phone. Every request carries
// an Idempotency-Key: a retry with the same key returns the order it took.
func (h *StaffOrderHandler) Create(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	actorID, actorRole, err := actorFromCtx(c)
	if err != nil {
		return err
	}
	checkoutKey, err := uuid.Parse(c.Get("Idempotency-Key"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("idempotency_key", "must be a UUID"))
	}
	var req dto.CreateStaffOrderRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.CreateOrder(c.Context(), actorID, actorRole, checkoutKey, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, out)
}

// Quote prices the items of an order about to be taken.
func (h *StaffOrderHandler) Quote(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	var req dto.StaffOrderQuoteRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Quote(c.Context(), req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// FindCustomer looks up the customer account with an email, to take an order
// for it.
func (h *StaffOrderHandler) FindCustomer(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	var req dto.StaffCustomerLookupRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.FindCustomer(c.Context(), req.Email)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
