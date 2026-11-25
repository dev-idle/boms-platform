package queue

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/port"
)

func TestReplyEmailQueue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rdb := newRedis(t)
	queue := NewEmailQueue(rdb)
	task := port.ReplyEmailTask{EventID: uuid.New(), ConversationID: uuid.New()}

	require.NoError(t, queue.EnqueueReplyEmail(ctx, task))
	require.NoError(t, queue.EnqueueReplyEmail(ctx, task), "the event delivered again is not an error")

	pending, err := asynq.NewInspectorFromRedisClient(rdb).ListPendingTasks(QueueEmail)
	require.NoError(t, err)
	require.Len(t, pending, 1, "and emails once")
	assert.Equal(t, TypeReplyEmail, pending[0].Type)
	assert.Equal(t, replyEmailMaxRetry, pending[0].MaxRetry)
	assert.Equal(t, replyEmailRetention, pending[0].Retention)

	decoded, err := decodeReplyEmail(asynq.NewTask(pending[0].Type, pending[0].Payload))
	require.NoError(t, err)
	assert.Equal(t, task, decoded)

	_, err = decodeReplyEmail(asynq.NewTask(TypeReplyEmail, []byte(`{"event_id":"x","conversation_id":"y"}`)))
	require.Error(t, err, "a task it cannot read is not retried")
}
