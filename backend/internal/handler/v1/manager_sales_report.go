package v1

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
)

// ManagerSalesReportHandler is a manager reading what the bakery sold.
type ManagerSalesReportHandler struct {
	usecase *usecase.ManagerSalesReportUsecase
}

func NewManagerSalesReportHandler(uc *usecase.ManagerSalesReportUsecase) *ManagerSalesReportHandler {
	return &ManagerSalesReportHandler{usecase: uc}
}

// Sales reports the bakery days ?from= to ?to= by ?group= (day, week or month).
func (h *ManagerSalesReportHandler) Sales(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.Report(c.Context(), c.Query("from"), c.Query("to"), c.Query("group"), time.Now())
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
