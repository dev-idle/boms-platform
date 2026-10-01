package middleware

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/infrastructure/ratelimit"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/response"
	"github.com/gofiber/fiber/v3"
	goredis "github.com/redis/go-redis/v9"
)

// rateLimitFailOpenCounter is incremented on fail-open paths (per-process).
var rateLimitFailOpenCounter atomic.Uint64

// RedisRateLimit returns middleware that rate-limits using a Redis sliding window.
// When failOpen is false, Redis errors return 503 (used for auth brute-force paths).
func RedisRateLimit(rdb *goredis.Client, keyFn func(fiber.Ctx) string, max int, window time.Duration, failOpen bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		if rdb == nil {
			return c.Next()
		}
		key := keyFn(c)
		if key == "" {
			return c.Next()
		}
		ctx := c.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		member := fmt.Sprintf("%d:%s", time.Now().UnixNano(), c.Get(fiber.HeaderXRequestID))
		ok, err := ratelimit.Allow(ctx, rdb, key, max, window, member)
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
		if ok {
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
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		return "rl:ip:" + ClientIP(c) + ":auth_attempt"
	}, cfg.AuthAttemptMax, cfg.AuthAttemptWindow, false)
}

// AuthLinkRateLimit limits following emailed links — confirming an address,
// setting a new password — per IP. It fails closed: a link token is too long
// to guess, and the cap keeps anyone from trying hard while Redis is away.
func AuthLinkRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		return "rl:ip:" + ClientIP(c) + ":auth_link"
	}, cfg.AuthLinkMax, cfg.AuthLinkWindow, false)
}

// PasswordResetRateLimit limits asking for password reset links per IP. It
// fails closed: each request may send an email to someone.
func PasswordResetRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		return "rl:ip:" + ClientIP(c) + ":password_reset"
	}, cfg.PasswordResetMax, cfg.PasswordResetWindow, false)
}

// AuthRefreshRateLimit limits refresh calls per IP.
func AuthRefreshRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		return "rl:ip:" + ClientIP(c) + ":auth_refresh"
	}, cfg.AuthRefreshMax, cfg.AuthRefreshWindow, true)
}

// AuthLogoutRateLimit limits logout calls per IP.
func AuthLogoutRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		return "rl:ip:" + ClientIP(c) + ":auth_logout"
	}, cfg.AuthLogoutMax, cfg.AuthLogoutWindow, true)
}

// AdminWriteRateLimit limits admin write operations per user.
func AdminWriteRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":admin_write"
		}
		return "rl:ip:" + ClientIP(c) + ":admin_write"
	}, cfg.AdminWriteMax, cfg.AdminWriteWindow, true)
}

// ManagerMediaRateLimit limits Cloudinary signature issuance per manager (separate from catalog writes).
func ManagerMediaRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":manager_media"
		}
		return "rl:ip:" + ClientIP(c) + ":manager_media"
	}, cfg.ManagerMediaMax, cfg.ManagerMediaWindow, true)
}

// ReferenceUploadRateLimit limits reference photo upload signatures per customer.
func ReferenceUploadRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":reference_upload"
		}
		return "rl:ip:" + ClientIP(c) + ":reference_upload"
	}, cfg.ReferenceUploadMax, cfg.ReferenceUploadWindow, true)
}

// ManagerWriteRateLimit limits manager catalog write operations per user.
func ManagerWriteRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":manager_write"
		}
		return "rl:ip:" + ClientIP(c) + ":manager_write"
	}, cfg.ManagerWriteMax, cfg.ManagerWriteWindow, true)
}

// SelfWriteRateLimit limits self-service account writes per user. PATCH /me answers
// phone_exists for a number another account holds and PATCH /me/password checks
// the old password, so both would otherwise be free oracles.
func SelfWriteRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":self_write"
		}
		return "rl:ip:" + ClientIP(c) + ":self_write"
	}, cfg.SelfWriteMax, cfg.SelfWriteWindow, true)
}

// DiscountAttemptRateLimit limits how often one customer may try a discount
// code. Every attempt answers whether the code exists, so without a cap the
// endpoint would let anyone walk the code space.
func DiscountAttemptRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":discount_attempt"
		}
		return "rl:ip:" + ClientIP(c) + ":discount_attempt"
	}, cfg.DiscountAttemptMax, cfg.DiscountAttemptWindow, true)
}

// RealtimeTicketRateLimit limits how often one user may ask for a realtime
// ticket. Each ticket costs a Redis write and, once redeemed, a socket.
func RealtimeTicketRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":realtime_ticket"
		}
		return "rl:ip:" + ClientIP(c) + ":realtime_ticket"
	}, cfg.RealtimeTicketMax, cfg.RealtimeTicketWindow, true)
}

// DataExportRateLimit limits how often one user may download their data. Each
// export reads every order they ever placed and hands out their personal data,
// so a stolen session cannot pull it in a loop.
func DataExportRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":data_export"
		}
		return "rl:ip:" + ClientIP(c) + ":data_export"
	}, cfg.DataExportMax, cfg.DataExportWindow, true)
}

// VerificationResendRateLimit limits how often one user asks for a new
// confirmation email.
func VerificationResendRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":verification_resend"
		}
		return "rl:ip:" + ClientIP(c) + ":verification_resend"
	}, cfg.VerificationResendMax, cfg.VerificationResendWindow, true)
}

// PaymentWebhookRateLimit limits payment provider notices per IP: checking
// each one costs a call to the provider, so a flood of forged ones is cut off
// before it makes that call.
func PaymentWebhookRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		return "rl:ip:" + ClientIP(c) + ":payment_webhook"
	}, cfg.PaymentWebhookMax, cfg.PaymentWebhookWindow, true)
}

// OrderWriteRateLimit limits order mutations (checkout, status transitions) per user.
func OrderWriteRateLimit(rdb *goredis.Client, cfg config.RateLimitRedisConfig) fiber.Handler {
	return RedisRateLimit(rdb, func(c fiber.Ctx) string {
		if uid, ok := GetUserID(c); ok {
			return "rl:user:" + uid.String() + ":order_write"
		}
		return "rl:ip:" + ClientIP(c) + ":order_write"
	}, cfg.OrderWriteMax, cfg.OrderWriteWindow, true)
}
