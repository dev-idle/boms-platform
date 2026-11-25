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
)

func TestWorkerHeartbeat(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	mr := miniredis.RunT(t)
	client, err := redisadapter.NewClient(ctx, config.RedisConfig{
		Addr: mr.Addr(), PoolSize: 4,
		DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second, HealthCheckTimeout: time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	heartbeat := redisadapter.NewWorkerHeartbeat(client)

	never, err := heartbeat.LastBeat(ctx)
	require.NoError(t, err)
	assert.Nil(t, never, "no run on record")

	at := time.Date(2026, 11, 25, 9, 30, 0, 0, time.UTC)
	require.NoError(t, heartbeat.Beat(ctx, at))
	last, err := heartbeat.LastBeat(ctx)
	require.NoError(t, err)
	require.NotNil(t, last)
	assert.True(t, last.Equal(at))

	mr.FastForward(8 * 24 * time.Hour)
	gone, err := heartbeat.LastBeat(ctx)
	require.NoError(t, err)
	assert.Nil(t, gone, "a run older than the record's life is not on record")
}
