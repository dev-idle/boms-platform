package middleware

import (
	"context"

	"github.com/boms/backend/internal/shared/ctxmeta"
	"github.com/gofiber/fiber/v3"
)

// AttachRequestMeta stores request IP and user agent in context for services.
func AttachRequestMeta() fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx := c.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		c.SetContext(ctxmeta.WithRequestMeta(ctx, ClientIP(c), c.Get(fiber.HeaderUserAgent)))
		return c.Next()
	}
}
