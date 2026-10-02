package queue

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/port"
)

func TestPromotionQueue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rdb := newRedis(t)
	queue := NewEmailQueue(rdb)
	inspector := asynq.NewInspectorFromRedisClient(rdb)
	promotion := port.PromotionTask{EventID: uuid.New(), PromotionID: uuid.New()}
	email := port.PromotionEmailTask{PromotionID: promotion.PromotionID, UserID: uuid.New()}

	require.NoError(t, queue.EnqueuePromotion(ctx, promotion))
	require.NoError(t, queue.EnqueuePromotion(ctx, promotion), "the event delivered again is not an error")
	require.NoError(t, queue.EnqueuePromotionEmail(ctx, email))
	require.NoError(t, queue.EnqueuePromotionEmail(ctx, email), "a promotion queued again mails nobody twice")

	pending, err := inspector.ListPendingTasks(QueuePromotion)
	require.NoError(t, err)
	require.Len(t, pending, 2, "in their own queue, behind the email customers wait for")
	types := []string{pending[0].Type, pending[1].Type}
	assert.ElementsMatch(t, []string{TypePromotion, TypePromotionEmail}, types)
	for _, task := range pending {
		assert.Equal(t, promotionMaxRetry, task.MaxRetry)
		assert.Equal(t, promotionRetention, task.Retention)
		assert.NotContains(t, string(task.Payload), "token", "a task never holds an unsubscribe link")
	}
}

type recordingPromotionSender struct {
	queued []port.PromotionTask
	sent   []port.PromotionEmailTask
	err    error
}

func (s *recordingPromotionSender) Queue(_ context.Context, task port.PromotionTask) error {
	s.queued = append(s.queued, task)
	return s.err
}

func (s *recordingPromotionSender) Send(_ context.Context, task port.PromotionEmailTask) error {
	s.sent = append(s.sent, task)
	return s.err
}

func TestPromotionHandlers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	task := func(t *testing.T, kind string, payload any) *asynq.Task {
		t.Helper()
		raw, err := json.Marshal(payload)
		require.NoError(t, err)
		return asynq.NewTask(kind, raw)
	}
	promotion := promotionPayload{EventID: uuid.NewString(), PromotionID: uuid.NewString()}
	email := promotionEmailPayload{PromotionID: promotion.PromotionID, UserID: uuid.NewString()}

	t.Run("hand_the_tasks_to_the_sender", func(t *testing.T) {
		t.Parallel()
		sender := &recordingPromotionSender{}
		require.NoError(t, PromotionHandler(sender)(ctx, task(t, TypePromotion, promotion)))
		require.NoError(t, PromotionEmailHandler(sender)(ctx, task(t, TypePromotionEmail, email)))
		require.Len(t, sender.queued, 1)
		assert.Equal(t, promotion.PromotionID, sender.queued[0].PromotionID.String())
		require.Len(t, sender.sent, 1)
		assert.Equal(t, email.UserID, sender.sent[0].UserID.String())
	})

	t.Run("an_undeliverable_email_is_not_retried", func(t *testing.T) {
		t.Parallel()
		sender := &recordingPromotionSender{err: port.ErrEmailUndeliverable}
		require.ErrorIs(t, PromotionEmailHandler(sender)(ctx, task(t, TypePromotionEmail, email)), asynq.SkipRetry)
	})

	t.Run("a_task_it_cannot_read_is_not_retried", func(t *testing.T) {
		t.Parallel()
		sender := &recordingPromotionSender{}
		require.ErrorIs(t, PromotionHandler(sender)(ctx, asynq.NewTask(TypePromotion, []byte("{"))), asynq.SkipRetry)
		require.ErrorIs(t, PromotionHandler(sender)(ctx, task(t, TypePromotion, promotionPayload{EventID: "x", PromotionID: promotion.PromotionID})), asynq.SkipRetry)
		require.ErrorIs(t, PromotionEmailHandler(sender)(ctx, task(t, TypePromotionEmail, promotionEmailPayload{PromotionID: email.PromotionID, UserID: "x"})), asynq.SkipRetry)
		assert.Empty(t, sender.queued)
		assert.Empty(t, sender.sent)
	})
}
