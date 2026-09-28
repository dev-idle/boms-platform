package v1

import (
	"testing"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Refresh cookie Path must stay "/" so the Next.js proxy can gate protected page routes.
// Narrow paths (e.g. /api/v1/auth) break post-login navigation — see frontend/src/proxy.ts.
func TestRefreshCookie_sitePathForProxyGate(t *testing.T) {
	t.Parallel()
	require.Equal(t, "/", middleware.AuthCookiePath, "AuthCookiePath contract with Next.js proxy")

	cfg := &config.Config{
		Cookie: config.CookieConfig{Name: "boms_refresh", Secure: false},
	}

	got := refreshCookie(cfg, "opaque-token", 900)
	require.NotNil(t, got)
	assert.Equal(t, "/", got.Path)
	assert.Equal(t, "boms_refresh", got.Name)
	assert.True(t, got.HTTPOnly)
	assert.Equal(t, fiber.CookieSameSiteLaxMode, got.SameSite)
	assert.Equal(t, 900, got.MaxAge)
}

func TestRefreshCookie_clearUsesSamePath(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Cookie: config.CookieConfig{Name: "boms_refresh"},
	}
	got := refreshCookie(cfg, "", -1)
	require.NotNil(t, got)
	assert.Equal(t, "/", got.Path)
	assert.Equal(t, -1, got.MaxAge)
}

func TestRefreshCookie_legacyPathConstant(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "/api/v1/auth", middleware.AuthCookieLegacyPath)
}

// The role cookie is what lets the Next.js proxy route a returning visitor at the
// edge. It must look exactly like the session cookie to the browser, and carry
// nothing but the role — no token, no identity.
func TestRoleCookie_matchesSessionCookieAttributes(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Cookie: config.CookieConfig{
			Name:     "boms_refresh",
			RoleName: "boms_role",
			Secure:   true,
		},
	}

	got := roleCookie(cfg, "manager", 900)
	require.NotNil(t, got)
	assert.Equal(t, "boms_role", got.Name)
	assert.Equal(t, "manager", got.Value)
	assert.Equal(t, middleware.AuthCookiePath, got.Path)
	assert.True(t, got.HTTPOnly, "a page-routing hint is still not for scripts")
	assert.True(t, got.Secure)
	assert.Equal(t, fiber.CookieSameSiteLaxMode, got.SameSite)
	assert.Equal(t, 900, got.MaxAge)
}

func TestRoleCookie_clearedWithTheSession(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		Cookie: config.CookieConfig{Name: "boms_refresh", RoleName: "boms_role"},
	}

	got := roleCookie(cfg, "", -1)
	require.NotNil(t, got)
	assert.Empty(t, got.Value)
	assert.Equal(t, -1, got.MaxAge, "a signed-out browser must not keep routing as the old role")
}
