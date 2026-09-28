package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/config"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/middleware"
)

// The browser only forgets a cookie when a Set-Cookie with the same name AND
// path expires it, so what matters is the headers that actually leave the
// server — not the cookie structs the handlers build.
func TestSessionCookies_Headers(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Cookie: config.CookieConfig{Name: "boms_refresh", RoleName: "boms_role", Secure: true},
		JWT:    config.JWTConfig{RefreshTTL: 15 * time.Minute},
	}

	cookiesFrom := func(t *testing.T, write func(c fiber.Ctx)) []*http.Cookie {
		t.Helper()
		app := fiber.New()
		app.Get("/", func(c fiber.Ctx) error {
			write(c)
			return c.SendStatus(fiber.StatusNoContent)
		})
		resp, err := app.Test(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil))
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		return resp.Cookies()
	}

	find := func(cookies []*http.Cookie, name, path string) *http.Cookie {
		for _, cookie := range cookies {
			if cookie.Name == name && cookie.Path == path {
				return cookie
			}
		}
		return nil
	}

	t.Run("clearing_expires_the_site_wide_refresh_cookie", func(t *testing.T) {
		cookies := cookiesFrom(t, func(c fiber.Ctx) { clearSessionCookies(c, cfg) })

		refresh := find(cookies, "boms_refresh", middleware.AuthCookiePath)
		require.NotNil(t, refresh, "logout must expire the refresh cookie the browser actually holds")
		assert.Empty(t, refresh.Value)
		assert.Negative(t, refresh.MaxAge)

		legacy := find(cookies, "boms_refresh", middleware.AuthCookieLegacyPath)
		require.NotNil(t, legacy, "the cookie an older build set under the auth path is expired too")
		assert.Negative(t, legacy.MaxAge)

		role := find(cookies, "boms_role", middleware.AuthCookiePath)
		require.NotNil(t, role)
		assert.Negative(t, role.MaxAge)
	})

	t.Run("signing_in_keeps_the_new_refresh_cookie_and_expires_the_legacy_one", func(t *testing.T) {
		cookies := cookiesFrom(t, func(c fiber.Ctx) {
			writeSessionCookies(c, cfg, "opaque-token", domainuser.RoleCustomer)
		})

		refresh := find(cookies, "boms_refresh", middleware.AuthCookiePath)
		require.NotNil(t, refresh)
		assert.Equal(t, "opaque-token", refresh.Value)
		assert.Positive(t, refresh.MaxAge)
		assert.True(t, refresh.HttpOnly)
		assert.True(t, refresh.Secure)

		legacy := find(cookies, "boms_refresh", middleware.AuthCookieLegacyPath)
		require.NotNil(t, legacy)
		assert.Negative(t, legacy.MaxAge)

		role := find(cookies, "boms_role", middleware.AuthCookiePath)
		require.NotNil(t, role)
		assert.Equal(t, string(domainuser.RoleCustomer), role.Value)
	})
}
