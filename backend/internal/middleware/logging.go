package middleware

import (
	"time"

	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// RequestLogger logs one line per request with latency, status, and correlation id.
// An error the handlers return is answered here, through the app's error
// handler, as Fiber's own logger does: until then the response still holds 200,
// and the line must record the status the client gets.
func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		chainErr := c.Next()
		if chainErr != nil {
			if err := c.App().ErrorHandler(c, chainErr); err != nil {
				// SendStatus only sets the status and body; there is no fallback past it.
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}

		status := c.Response().StatusCode()
		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Duration("latency", time.Since(start)),
			zap.String("remote_ip", ClientIP(c)),
			zap.Int("bytes_out", len(c.Response().Body())),
		}
		fields = append(fields, response.ZapCorrelationFields(c)...)
		if ua := c.Get(fiber.HeaderUserAgent); ua != "" {
			fields = append(fields, zap.String("user_agent", ua))
		}
		if chainErr != nil {
			fields = append(fields, zap.Error(chainErr))
		}
		switch {
		case status >= fiber.StatusInternalServerError:
			log.Error("http_request", fields...)
		case chainErr != nil:
			log.Warn("http_request", fields...)
		default:
			log.Info("http_request", fields...)
		}
		return nil
	}
}
