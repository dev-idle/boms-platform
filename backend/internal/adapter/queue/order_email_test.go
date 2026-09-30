package queue

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

func newRedis(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: miniredis.RunT(t).Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

func TestOrderEmailQueue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("one_task_per_event", func(t *testing.T) {
		t.Parallel()
		rdb := newRedis(t)
		queue := NewOrderEmailQueue(rdb)
		task := port.OrderEmailTask{EventID: uuid.New(), OrderID: uuid.New(), Notice: domainorder.NoticeReady}

		require.NoError(t, queue.EnqueueOrderEmail(ctx, task))
		require.NoError(t, queue.EnqueueOrderEmail(ctx, task), "the event delivered again is not an error")

		pending, err := asynq.NewInspectorFromRedisClient(rdb).ListPendingTasks(QueueEmail)
		require.NoError(t, err)
		require.Len(t, pending, 1, "and does not mail the customer twice")
		assert.Equal(t, task.EventID.String(), pending[0].ID)
		assert.Equal(t, TypeOrderEmail, pending[0].Type)
		assert.Equal(t, orderEmailMaxRetry, pending[0].MaxRetry)
		assert.Equal(t, orderEmailRetention, pending[0].Retention, "the window a redelivered event is recognised in")
		assert.Equal(t, orderEmailTimeout, pending[0].Timeout)

		decoded, err := decodeOrderEmail(asynq.NewTask(pending[0].Type, pending[0].Payload))
		require.NoError(t, err)
		assert.Equal(t, task, decoded)
	})

	t.Run("a_redis_failure_is_reported", func(t *testing.T) {
		t.Parallel()
		server := miniredis.RunT(t)
		rdb := redis.NewClient(&redis.Options{Addr: server.Addr()})
		t.Cleanup(func() { _ = rdb.Close() })
		server.Close()

		err := NewOrderEmailQueue(rdb).EnqueueOrderEmail(ctx, port.OrderEmailTask{
			EventID: uuid.New(), OrderID: uuid.New(), Notice: domainorder.NoticePlaced,
		})

		require.Error(t, err, "the outbox keeps the event and delivers it again")
	})
}

type recordingSender struct {
	got []port.OrderEmailTask
	err error
}

func (s *recordingSender) Send(_ context.Context, task port.OrderEmailTask) error {
	s.got = append(s.got, task)
	return s.err
}

func TestOrderEmailHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	payload := func(t *testing.T, p orderEmailPayload) *asynq.Task {
		t.Helper()
		raw, err := json.Marshal(p)
		require.NoError(t, err)
		return asynq.NewTask(TypeOrderEmail, raw)
	}
	valid := orderEmailPayload{EventID: uuid.NewString(), OrderID: uuid.NewString(), Notice: string(domainorder.NoticePlaced)}

	t.Run("hands_the_task_to_the_sender", func(t *testing.T) {
		t.Parallel()
		sender := &recordingSender{}

		require.NoError(t, OrderEmailHandler(sender)(ctx, payload(t, valid)))

		require.Len(t, sender.got, 1)
		assert.Equal(t, valid.OrderID, sender.got[0].OrderID.String())
		assert.Equal(t, valid.EventID, sender.got[0].EventID.String())
		assert.Equal(t, domainorder.NoticePlaced, sender.got[0].Notice)
	})

	t.Run("a_passing_failure_is_retried", func(t *testing.T) {
		t.Parallel()
		down := errors.New("mail server unreachable")

		err := OrderEmailHandler(&recordingSender{err: down})(ctx, payload(t, valid))

		require.ErrorIs(t, err, down)
		assert.NotErrorIs(t, err, asynq.SkipRetry)
	})

	t.Run("an_undeliverable_email_is_not_retried", func(t *testing.T) {
		t.Parallel()
		sender := &recordingSender{err: port.ErrEmailUndeliverable}

		err := OrderEmailHandler(sender)(ctx, payload(t, valid))

		require.ErrorIs(t, err, asynq.SkipRetry)
	})

	t.Run("a_task_it_cannot_read_is_not_retried", func(t *testing.T) {
		t.Parallel()
		bad := []*asynq.Task{
			asynq.NewTask(TypeOrderEmail, []byte("{")),
			payload(t, orderEmailPayload{EventID: valid.EventID, OrderID: "x", Notice: valid.Notice}),
			payload(t, orderEmailPayload{EventID: valid.EventID, OrderID: valid.OrderID, Notice: "shipped"}),
			payload(t, orderEmailPayload{EventID: "x", OrderID: valid.OrderID, Notice: valid.Notice}),
		}
		sender := &recordingSender{}
		for _, task := range bad {
			require.ErrorIs(t, OrderEmailHandler(sender)(ctx, task), asynq.SkipRetry, string(task.Payload()))
		}
		assert.Empty(t, sender.got)
	})
}
