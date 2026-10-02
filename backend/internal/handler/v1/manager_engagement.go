package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
)

// ManagerEngagementHandler is a manager reading the engagement report.
type ManagerEngagementHandler struct {
	usecase *usecase.ManagerEngagementUsecase
}

func NewManagerEngagementHandler(uc *usecase.ManagerEngagementUsecase) *ManagerEngagementHandler {
	return &ManagerEngagementHandler{usecase: uc}
}

// Report returns how customers use the engagement features.
func (h *ManagerEngagementHandler) Report(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.Report(c.Context())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
