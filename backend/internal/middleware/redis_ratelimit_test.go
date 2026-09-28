package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/config"
)

// Per-user limiters that fail open: the discount limiter, where each answer says
// whether a code exists so the cap stops the code space being walked, and the
// realtime ticket limiter, where each ticket costs a Redis write and a socket.
func TestPerUserRateLimits(t *testing.T) {
	t.Parallel()

	cfg := config.RateLimitRedisConfig{
		DiscountAttemptMax: 2, DiscountAttemptWindow: time.Minute,
		RealtimeTicketMax: 2, RealtimeTicketWindow: time.Minute,
	}
	limiters := map[string]func(*goredis.Client, config.RateLimitRedisConfig) fiber.Handler{
		"discount_attempts": DiscountAttemptRateLimit,
		"realtime_tickets":  RealtimeTicketRateLimit,
	}
	for name, limiter := range limiters {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testPerUserRateLimit(t, func(rdb *goredis.Client) fiber.Handler { return limiter(rdb, cfg) })
		})
	}
}

func testPerUserRateLimit(t *testing.T, limiter func(*goredis.Client) fiber.Handler) {
	t.Helper()

	newApp := func(rdb *goredis.Client) *fiber.App {
		app := fiber.New()
		app.Post("/limited", func(c fiber.Ctx) error {
			if id := c.Get("X-Test-User"); id != "" {
				c.Locals(localUserIDKey, uuid.MustParse(id))
			}
			return c.Next()
		}, limiter(rdb), func(c fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		return app
	}

	// attempt returns the status and headers of one try; the body is not needed.
	attempt := func(t *testing.T, app *fiber.App, userID uuid.UUID) (int, http.Header) {
		t.Helper()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/limited", nil)
		req.Header.Set("X-Test-User", userID.String())
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode, resp.Header
	}

	t.Run("refuses_attempts_past_the_cap_with_a_retry_hint", func(t *testing.T) {
		t.Parallel()
		app := newApp(goredis.NewClient(&goredis.Options{Addr: miniredis.RunT(t).Addr()}))
		user := uuid.New()

		for range 2 {
			status, _ := attempt(t, app, user)
			assert.Equal(t, fiber.StatusOK, status)
		}
		status, header := attempt(t, app, user)
		assert.Equal(t, fiber.StatusTooManyRequests, status)
		assert.Equal(t, "60", header.Get(fiber.HeaderRetryAfter))
	})

	t.Run("keeps_a_separate_count_per_user", func(t *testing.T) {
		t.Parallel()
		app := newApp(goredis.NewClient(&goredis.Options{Addr: miniredis.RunT(t).Addr()}))

		first, second := uuid.New(), uuid.New()
		attempt(t, app, first)
		attempt(t, app, first)
		status, _ := attempt(t, app, second)
		assert.Equal(t, fiber.StatusOK, status)
	})

	t.Run("lets_the_attempt_through_when_redis_is_down", func(t *testing.T) {
		t.Parallel()
		server := miniredis.RunT(t)
		app := newApp(goredis.NewClient(&goredis.Options{Addr: server.Addr(), MaxRetries: -1}))
		server.Close()

		status, header := attempt(t, app, uuid.New())
		assert.Equal(t, fiber.StatusOK, status)
		assert.Equal(t, "1", header.Get("X-RateLimit-FailOpen"))
	})
}
