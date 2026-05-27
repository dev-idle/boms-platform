package middleware

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/boms/backend/internal/config"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	goredis "github.com/redis/go-redis/v9"
)

// rateLimitFailOpenCounter is incremented on fail-open paths (per-process).
var rateLimitFailOpenCounter atomic.Uint64

// slidingWindowScript implements a Redis sorted-set sliding window rate limiter.
var slidingWindowScript = goredis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local member = ARGV[4]
redis.call('ZREMRANGEBYSCORE', key, 0, now - window_ms)
if redis.call('ZCARD', key) >= limit then
  return 0
end
redis.call('ZADD', key, now, member)
redis.call('PEXPIRE', key, window_ms)
return 1
`)

// RedisRateLimit returns middleware that rate-limits using a Redis sliding window.
// When failOpen is false, Redis errors return 503 (used for auth brute-force paths).
func RedisRateLimit(rdb *goredis.Client, keyFn func(*fiber.Ctx) string, max int, window time.Duration, failOpen bool) fiber.Handler {
	windowMs := window.Milliseconds()
	if windowMs < 1 {
		windowMs = 1000
	}
	return func(c *fiber.Ctx) error {
		if rdb == nil {
			return c.Next()
		}
		key := keyFn(c)
		if key == "" {
			return c.Next()
		}
		ctx := c.UserContext()
		if ctx == nil {
			ctx = context.Background()
		}
		now := time.Now()
		member := fmt.Sprintf("%d:%s", now.UnixNano(), c.Get(fiber.HeaderXRequestID))
		ok, err := slidingWindowScript.Run(ctx, rdb, []string{key},
			max, windowMs, now.UnixMilli(), member,
		).Int()
		if err != nil {
			if !failOpen {
				return response.Error(c, fiber.StatusServiceUnavailable, &response.ErrorBody{
					Code:    apperrors.ErrServiceUnavailable.Code,
					Message: apperrors.ErrServiceUnavailable.Message,
				})
			}
			rateLimitFailOpenCounter.Add(1)
			c.Set("X-RateLimit-FailOpen", "1")
			return c.Next()
		}
		if ok == 1 {
			return c.Next()
		}
		sec := int(window.Round(time.Second).Seconds())
		if sec < 1 {
			sec = 1
		}
		c.Set(fiber.HeaderRetryAfter, strconv.Itoa(sec))
		return response.Error(c, fiber.StatusTooManyRequests, &response.ErrorBody{
			Code:    apperrors.ErrTooManyRequests.Code,
			Message: apperrors.ErrTooManyRequests.Message,
		})
	}
}

// AuthAttemptRateLimit limits login/register attempts per IP (fail-closed when Redis is down).
func AuthAttemptRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c *fiber.Ctx) string {
		return "rl:ip:" + c.IP() + ":auth_attempt"
	}, cfg.AuthAttemptMax, cfg.AuthAttemptWindow, false)
}

// AuthRefreshRateLimit limits refresh calls per IP.
func AuthRefreshRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c *fiber.Ctx) string {
		return "rl:ip:" + c.IP() + ":auth_refresh"
	}, cfg.AuthRefreshMax, cfg.AuthRefreshWindow, true)
}

// AuthLogoutRateLimit limits logout calls per IP.
func AuthLogoutRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c *fiber.Ctx) string {
		return "rl:ip:" + c.IP() + ":auth_logout"
	}, cfg.AuthLogoutMax, cfg.AuthLogoutWindow, true)
}

// AuthUserRateLimit limits authenticated routes per user id (optional; wire on sensitive authed endpoints).
func AuthUserRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c *fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":auth"
		}
		return "rl:ip:" + c.IP() + ":auth"
	}, cfg.AuthUserMax, cfg.AuthUserWindow, true)
}

// AdminWriteRateLimit limits admin write operations per user.
func AdminWriteRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c *fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":admin_write"
		}
		return "rl:ip:" + c.IP() + ":admin_write"
	}, cfg.AdminWriteMax, cfg.AdminWriteWindow, true)
}
