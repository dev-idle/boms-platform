package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
)

// PaymentHandler serves paying for an order through PayPal and PayPal's
// notices about payments.
type PaymentHandler struct {
	usecase *usecase.PaymentUsecase
}

func NewPaymentHandler(uc *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{usecase: uc}
}

// Start handles POST /api/v1/orders/:id/payment.
func (h *PaymentHandler) Start(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	approveURL, err := h.usecase.Start(c.Context(), userID, orderID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, dto.PaymentStartResponse{ApproveURL: approveURL})
}

// Capture handles POST /api/v1/orders/:id/payment/capture.
func (h *PaymentHandler) Capture(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	status, err := h.usecase.Capture(c.Context(), userID, orderID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, dto.PaymentCaptureResponse{Status: string(status)})
}

// Webhook handles POST /api/v1/payments/paypal/webhook: PayPal's signed
// notices, which carry no session.
func (h *PaymentHandler) Webhook(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	header := func(name string) string { return c.Get(name) }
	if err := h.usecase.HandleWebhook(c.Context(), header, c.Body()); err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.NoContent(c)
}
