package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"github.com/boms/backend/internal/config"
)

// CORS returns Fiber CORS middleware. Origins must be explicitly configured;
// wildcard is only allowed when AllowCredentials is false (CORS spec).
// The browser rejects credentialed requests against "*" anyway — keep the server fail-closed.
func CORS(cfg config.CORSConfig) fiber.Handler {
	origins := cfg.AllowOrigins
	if len(origins) == 0 {
		if cfg.AllowCredentials {
			// No origin configured and credentials requested — block all cross-origin requests.
			return func(c fiber.Ctx) error { return c.Next() }
		}
		origins = []string{"*"}
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
