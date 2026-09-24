package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestTimeout gives every handler a deadline.
//
// Fiber's WriteTimeout closes the socket but does not stop the handler: a client
// that walks away leaves the work running, still holding its database
// connections. The deadline travels on the user context, so the queries a
// handler has in flight are cancelled with it. Keep it below WriteTimeout, so
// the handler gives up first and the client gets the mapped error rather than a
// dropped connection.
func RequestTimeout(d time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if d <= 0 {
			return c.Next()
		}

		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		ctx, cancel := context.WithTimeout(ctx, d)
		defer cancel()

		c.SetUserContext(ctx)
		return c.Next()
	}
}
