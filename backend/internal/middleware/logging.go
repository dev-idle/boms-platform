package middleware

import (
	"time"

	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RequestLogger logs one line per request with latency, status, and correlation id.
func RequestLogger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		fields := []zap.Field{
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", time.Since(start)),
			zap.String("remote_ip", c.IP()),
			zap.Int("bytes_out", len(c.Response().Body())),
		}
		if rid := response.RequestIDFromCtx(c); rid != "" {
			fields = append(fields, zap.String("request_id", rid))
		}
		if ua := c.Get(fiber.HeaderUserAgent); ua != "" {
			fields = append(fields, zap.String("user_agent", ua))
		}
		if err != nil && c.Response().StatusCode() < 500 {
			fields = append(fields, zap.Error(err))
			log.Warn("http_request", fields...)
			return err
		}
		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Error("http_request", fields...)
			return err
		}

		log.Info("http_request", fields...)
		return nil
	}
}
