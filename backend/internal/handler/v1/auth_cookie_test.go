package v1

import (
	"testing"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
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
