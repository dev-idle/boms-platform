package v1

import (
	"github.com/gofiber/fiber/v3"

	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/boms/backend/internal/usecase"
)

// ManagerIncidentHandler is a manager reading the incident log.
type ManagerIncidentHandler struct {
	usecase *usecase.ManagerIncidentUsecase
}

func NewManagerIncidentHandler(uc *usecase.ManagerIncidentUsecase) *ManagerIncidentHandler {
	return &ManagerIncidentHandler{usecase: uc}
}

// List pages the incidents of the week ?week= starts, latest first; ?type=
// narrows them.
func (h *ManagerIncidentHandler) List(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	page := utils.ParseQueryInt32(c.Query("page", "1"), 1)
	pageSize := utils.ParseQueryInt32(
		c.Query("page_size", usecase.CatalogListDefaultPageSizeQuery),
		usecase.CatalogListDefaultPageSize,
	)
	out, total, page, pageSize, err := h.usecase.List(c.Context(), c.Query("week"), c.Query("type"), page, pageSize)
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OKPaginated(c, out, int(page), int(pageSize), total)
}

// Summary counts the incidents of each type in the week ?week= starts.
func (h *ManagerIncidentHandler) Summary(c fiber.Ctx) error {
	response.EnsureRequestID(c)
	out, err := h.usecase.Summary(c.Context(), c.Query("week"))
	if err != nil {
		return writeMapUsecaseError(c, err)
	}
	return response.OK(c, out)
}
