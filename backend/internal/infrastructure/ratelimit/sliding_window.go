// Package ratelimit counts uses of a key over a sliding window in Redis.
package ratelimit

import (
	"context"
	"time"

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

// Allow records one use of key, named member (unique per use), and reports
// whether it is within max uses in the last window. A refused use is not
// recorded, so a caller that keeps trying is let in again as the window moves.
func Allow(ctx context.Context, rdb *goredis.Client, key string, max int, window time.Duration, member string) (bool, error) {
	windowMs := window.Milliseconds()
	if windowMs < 1 {
		windowMs = 1000
	}
	ok, err := slidingWindowScript.Run(ctx, rdb, []string{key}, max, windowMs, time.Now().UnixMilli(), member).Int()
	if err != nil {
		return false, err
	}
	return ok == 1, nil
}
