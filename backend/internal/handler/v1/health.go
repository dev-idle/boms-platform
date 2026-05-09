package v1

import (
	"github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/boms/backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

// HealthHandler exposes operational endpoints (no domain logic).
type HealthHandler struct {
	readiness *usecase.Readiness
}

// NewHealthHandler wires readiness checks for orchestration probes.
func NewHealthHandler(readiness *usecase.Readiness) *HealthHandler {
	return &HealthHandler{readiness: readiness}
}

// Live handles GET /health (process liveness).
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	return response.OK(c, fiber.Map{
		"status": "ok",
	})
}

// Ready handles GET /ready (dependency readiness).
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	body := h.readiness.Execute(c.UserContext())
	if body.Status != "ready" {
		return response.JSON(c, fiber.StatusServiceUnavailable, false, body, &response.ErrorBody{
			Code:    errors.ErrServiceUnavailable.Code,
			Message: "One or more dependencies are not ready",
		}, response.MetaFromCtx(c, nil))
	}
	return response.OK(c, body)
}
