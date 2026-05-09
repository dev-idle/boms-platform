package middleware

import (
	"fmt"

	"github.com/boms/backend/internal/config"
	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders sets baseline hardening headers (2026 OWASP-style defaults for JSON APIs).
// HSTS is optional: set HTTP_HSTS_MAX_AGE>0 only when clients always reach this app over HTTPS
// (typically at the TLS-terminating reverse proxy — still safe to emit at app if appropriate).
func SecurityHeaders(cfg config.HTTPConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
		c.Set(fiber.HeaderXFrameOptions, "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")
		if cfg.HSTSMaxAge > 0 {
			c.Set(fiber.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d", cfg.HSTSMaxAge))
		}
		return c.Next()
	}
}
