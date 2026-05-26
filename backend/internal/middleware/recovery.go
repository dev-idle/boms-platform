package middleware

import (
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Recover catches panics, logs a stack trace, and returns a generic 500 JSON envelope.
// Register this as early as possible in the middleware chain so it catches panics
// from downstream middleware (security headers, logging, cors, rate limit).
func Recover(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				fields := []zap.Field{
					zap.Any("panic", r),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.Stack("stack"),
				}
				fields = append(fields, response.ZapCorrelationFields(c)...)
				log.Error("panic_recovered", fields...)
				response.EnsureRequestID(c)
				_ = response.Error(c, apperrors.ErrInternal.StatusCode, &response.ErrorBody{
					Code:    apperrors.ErrInternal.Code,
					Message: apperrors.ErrInternal.Message,
				})
			}
		}()
		return c.Next()
	}
}
