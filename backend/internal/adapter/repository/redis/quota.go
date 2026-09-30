package redis

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/boms/backend/internal/infrastructure/ratelimit"
	"github.com/boms/backend/internal/port"
)

const quotaKeyPrefix = "quota:"

// Quota counts uses per key over a sliding window, like the request limiters;
// the key expires with the window, so nothing is left to clean up.
type Quota struct {
	rdb *redis.Client
}

func NewQuota(client *Client) *Quota {
	return &Quota{rdb: client.RDB()}
}

// Take implements port.Quota.
func (q *Quota) Take(ctx context.Context, key string, limit port.QuotaLimit) (bool, error) {
	ok, err := ratelimit.Allow(ctx, q.rdb, quotaKeyPrefix+key, limit.Max, limit.Window, uuid.NewString())
	if err != nil {
		return false, fmt.Errorf("take quota: %w", err)
	}
	return ok, nil
}

var _ port.Quota = (*Quota)(nil)
