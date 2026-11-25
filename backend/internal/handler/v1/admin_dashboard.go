package v1

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
)

// AdminDashboardHandler serves the admin home.
type AdminDashboardHandler struct {
	usecase *usecase.AdminDashboardUsecase
}

func NewAdminDashboardHandler(uc *usecase.AdminDashboardUsecase) *AdminDashboardHandler {
	return &AdminDashboardHandler{usecase: uc}
}

// Get handles GET /api/v1/admin/dashboard.
func (h *AdminDashboardHandler) Get(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.Get(c.Context(), time.Now())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
