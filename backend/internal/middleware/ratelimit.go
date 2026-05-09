package middleware

import (
	"strconv"
	"time"

	"github.com/boms/backend/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimit returns a simple in-memory per-IP rate limiter.
func RateLimit(cfg config.RateLimitConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Max,
		Expiration: cfg.WindowDuration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			sec := int(cfg.WindowDuration.Round(time.Second).Seconds())
			if sec < 1 {
				sec = 1
			}
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(sec))
			return fiber.ErrTooManyRequests
		},
		Next: func(c *fiber.Ctx) bool {
			switch c.Path() {
			case "/health", "/ready":
				return true
			default:
				return false
			}
		},
	})
}
