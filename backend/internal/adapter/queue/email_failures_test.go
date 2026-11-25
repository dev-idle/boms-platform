package queue

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An email the worker gave up on is archived in its queue; a queue no email
// has gone through yet counts none.
func TestEmailFailures(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rdb := newRedis(t)
	failures := NewEmailFailures(rdb)

	none, err := failures.Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, none)

	client := asynq.NewClientFromRedisClient(rdb)
	for _, name := range []string{QueueEmail, QueuePromotion} {
		info, err := client.Enqueue(asynq.NewTask("test", nil), asynq.Queue(name))
		require.NoError(t, err)
		require.NoError(t, asynq.NewInspectorFromRedisClient(rdb).ArchiveTask(name, info.ID))
	}

	failed, err := failures.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), failed)
}
