package response

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ZapCorrelationFields returns structured log fields for request/trace correlation.
func ZapCorrelationFields(c *fiber.Ctx) []zap.Field {
	id := RequestIDFromCtx(c)
	if id == "" {
		return nil
	}
	return []zap.Field{
		zap.String("request_id", id),
		zap.String("trace_id", id),
	}
}
