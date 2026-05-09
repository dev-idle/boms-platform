package middleware

import (
	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Recover catches panics, logs a stack trace, and returns a generic 500 JSON envelope.
func Recover(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic_recovered",
					zap.Any("panic", r),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.Stack("stack"),
				)
				response.EnsureRequestID(c)
				_ = response.Error(c, fiber.StatusInternalServerError, &response.ErrorBody{
					Code:    "internal_error",
					Message: "An unexpected error occurred",
				})
			}
		}()
		return c.Next()
	}
}
