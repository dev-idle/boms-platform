package middleware

import (
	"strings"

	"github.com/boms/backend/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS returns Fiber CORS middleware. Origins must be explicitly configured;
// wildcard is only allowed when AllowCredentials is false (CORS spec).
// The browser rejects credentialed requests against "*" anyway — keep the server fail-closed.
func CORS(cfg config.CORSConfig) fiber.Handler {
	origins := strings.Join(cfg.AllowOrigins, ",")
	if origins == "" {
		if cfg.AllowCredentials {
			// No origin configured and credentials requested — block all cross-origin requests.
			return func(c *fiber.Ctx) error { return c.Next() }
		}
		origins = "*"
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Authorization,Content-Type,X-Request-ID",
		ExposeHeaders:    "X-Request-ID",
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
