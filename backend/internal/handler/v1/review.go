package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	sharevalidator "github.com/boms/backend/internal/shared/validator"
	"github.com/boms/backend/internal/usecase"
)

// ReviewHandler is customers reviewing what they picked up, and the
// storefront showing the published reviews.
type ReviewHandler struct {
	usecase *usecase.ReviewUsecase
}

func NewReviewHandler(uc *usecase.ReviewUsecase) *ReviewHandler {
	return &ReviewHandler{usecase: uc}
}

// ListByOrder returns the customer's reviews of the products on their order.
func (h *ReviewHandler) ListByOrder(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	out, err := h.usecase.ListByOrder(c.Context(), userID, orderID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// Create records the customer's review of a product on their order.
func (h *ReviewHandler) Create(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	orderID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid order id"))
	}
	var req dto.CreateReviewRequest
	if err := c.Bind().Body(&req); err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("body", "invalid request body"))
	}
	if err := sharevalidator.Struct(&req); err != nil {
		return writeValidationError(c, err)
	}
	out, err := h.usecase.Create(c.Context(), userID, orderID, req)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.Created(c, out)
}

// ProductReviews returns a page of a product's published reviews
// (?before=<review id> for older ones).
func (h *ReviewHandler) ProductReviews(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	productID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return writeAppError(c, apperrors.ErrValidation.WithDetail("id", "invalid product id"))
	}
	beforeID, appErr := parseBeforeID(c, "review")
	if appErr != nil {
		return writeAppError(c, appErr)
	}
	out, err := h.usecase.ProductReviews(c.Context(), productID, beforeID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
