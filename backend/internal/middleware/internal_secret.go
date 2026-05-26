package middleware

import (
	"crypto/subtle"

	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/gofiber/fiber/v2"
)

// InternalSecretHeader is the header name used between the Next.js proxy and this API.
const InternalSecretHeader = "X-Internal-Secret" //nolint:gosec // header name, not a credential

// RequireInternalSecret enforces a shared secret on every request. The Next.js proxy
// signs each forwarded request with the configured secret; clients hitting the API
// directly (without going through the proxy) are rejected with 401.
//
// When the configured secret is empty the middleware is a no-op — convenient for
// local development. Production config validation requires the secret to be set,
// so this only short-circuits in dev environments.
//
// Comparison is constant-time to avoid timing side-channels.
func RequireInternalSecret(secret string) fiber.Handler {
	if secret == "" {
		return func(c *fiber.Ctx) error { return c.Next() }
	}
	expected := []byte(secret)
	return func(c *fiber.Ctx) error {
		got := []byte(c.Get(InternalSecretHeader))
		if len(got) == 0 || subtle.ConstantTimeCompare(got, expected) != 1 {
			return writeMiddlewareError(c, apperrors.ErrUnauthorized)
		}
		return c.Next()
	}
}
