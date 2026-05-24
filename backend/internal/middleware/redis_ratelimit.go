package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v2"
	goredis "github.com/redis/go-redis/v9"
)

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
// On Redis errors the request is allowed through (fail-open) so auth remains available during cache outages.
func RedisRateLimit(rdb *goredis.Client, keyFn func(*fiber.Ctx) string, max int, window time.Duration) fiber.Handler {
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

// AuthAttemptRateLimit limits login/register attempts per IP.
func AuthAttemptRateLimit(rdb *goredis.Client) fiber.Handler {
	return RedisRateLimit(rdb, func(c *fiber.Ctx) string {
		return "rl:ip:" + c.IP() + ":auth_attempt"
	}, 5, time.Minute)
}

// AuthRefreshRateLimit limits refresh calls per IP (10/min — balances abuse prevention vs legit polling).
func AuthRefreshRateLimit(rdb *goredis.Client) fiber.Handler {
	return RedisRateLimit(rdb, func(c *fiber.Ctx) string {
		return "rl:ip:" + c.IP() + ":auth_refresh"
	}, 10, time.Minute)
}

// AuthLogoutRateLimit limits logout calls per IP.
func AuthLogoutRateLimit(rdb *goredis.Client) fiber.Handler {
	return RedisRateLimit(rdb, func(c *fiber.Ctx) string {
		return "rl:ip:" + c.IP() + ":auth_logout"
	}, 10, time.Minute)
}

// AuthUserRateLimit limits authenticated routes per user id (optional; use on sensitive authed endpoints).
func AuthUserRateLimit(rdb *goredis.Client) fiber.Handler {
	return RedisRateLimit(rdb, func(c *fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":auth"
		}
		return "rl:ip:" + c.IP() + ":auth"
	}, 60, time.Minute)
}
