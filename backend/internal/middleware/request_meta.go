package middleware

import (
	"context"

	"github.com/boms/backend/internal/shared/ctxmeta"
	"github.com/gofiber/fiber/v2"
)

// AttachRequestMeta stores request IP and user agent in context for services.
func AttachRequestMeta() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		c.SetUserContext(ctxmeta.WithRequestMeta(ctx, c.IP(), c.Get(fiber.HeaderUserAgent)))
		return c.Next()
	}
}
