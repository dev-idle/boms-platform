package v1

import (
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

// Ready handles GET /ready (dependency readiness). Returns sanitized JSON only.
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	response.EnsureRequestID(c)
	body := h.readiness.Execute(c.UserContext())
	if body.Status != "ok" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(body)
	}
	return c.Status(fiber.StatusOK).JSON(body)
}
