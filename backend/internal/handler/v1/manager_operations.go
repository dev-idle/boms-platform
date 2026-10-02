package v1

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
)

// ManagerOperationsHandler is a manager watching the bakery's work as it goes.
type ManagerOperationsHandler struct {
	usecase *usecase.ManagerOperationsUsecase
}

func NewManagerOperationsHandler(uc *usecase.ManagerOperationsUsecase) *ManagerOperationsHandler {
	return &ManagerOperationsHandler{usecase: uc}
}

// Dashboard returns where the bakery's work stands now.
func (h *ManagerOperationsHandler) Dashboard(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.Dashboard(c.Context(), time.Now())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
