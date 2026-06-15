package v1

import (
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type ManagerMediaHandler struct {
	usecase *usecase.ManagerMediaUsecase
}

func NewManagerMediaHandler(uc *usecase.ManagerMediaUsecase) *ManagerMediaHandler {
	return &ManagerMediaHandler{usecase: uc}
}

func (h *ManagerMediaHandler) GetCloudinaryUploadSignature(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.CloudinaryUploadSignature()
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
