package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/middleware"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
)

type MediaHandler struct {
	usecase *usecase.MediaUsecase
}

func NewMediaHandler(uc *usecase.MediaUsecase) *MediaHandler {
	return &MediaHandler{usecase: uc}
}

// ProductImageSignature signs a manager's catalog image upload.
func (h *MediaHandler) ProductImageSignature(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.ProductImageSignature()
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}

// ReferenceImageSignature signs a customer's reference photo upload.
func (h *MediaHandler) ReferenceImageSignature(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return writeAppError(c, apperrors.ErrUnauthorized)
	}
	out, err := h.usecase.ReferenceImageSignature(c.Context(), userID)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
