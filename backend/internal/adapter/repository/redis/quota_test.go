package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redisadapter "github.com/boms/backend/internal/adapter/repository/redis"
	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
)

func TestQuota(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mr := miniredis.RunT(t)
	client, err := redisadapter.NewClient(ctx, config.RedisConfig{
		Addr: mr.Addr(), PoolSize: 4,
		DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second, HealthCheckTimeout: time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	quota := redisadapter.NewQuota(client)
	limit := port.QuotaLimit{Max: 2, Window: time.Hour}

	for i := range 2 {
		ok, err := quota.Take(ctx, "password_reset:u1", limit)
		require.NoError(t, err)
		assert.True(t, ok, "use %d is within the limit", i+1)
	}
	ok, err := quota.Take(ctx, "password_reset:u1", limit)
	require.NoError(t, err)
	assert.False(t, ok, "a third use within the hour is refused")

	ok, err = quota.Take(ctx, "password_reset:u2", limit)
	require.NoError(t, err)
	assert.True(t, ok, "another key has its own uses")

	mr.FastForward(time.Hour + time.Second)
	ok, err = quota.Take(ctx, "password_reset:u1", limit)
	require.NoError(t, err)
	assert.True(t, ok, "the uses come back when the window ends")

	mr.Close()
	_, err = quota.Take(ctx, "password_reset:u3", limit)
	assert.Error(t, err, "a Redis failure is reported, not taken as permission")
}
