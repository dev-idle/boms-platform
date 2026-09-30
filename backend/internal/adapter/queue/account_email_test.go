package queue

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainaccount "github.com/boms/backend/internal/domain/account"
	"github.com/boms/backend/internal/port"
)

func TestAccountEmailQueue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	rdb := newRedis(t)
	queue := NewEmailQueue(rdb)
	task := port.AccountEmailTask{EventID: uuid.New(), UserID: uuid.New(), Purpose: domainaccount.PurposeResetPassword}

	require.NoError(t, queue.EnqueueAccountEmail(ctx, task))
	require.NoError(t, queue.EnqueueAccountEmail(ctx, task), "the event delivered again is not an error")

	pending, err := asynq.NewInspectorFromRedisClient(rdb).ListPendingTasks(QueueEmail)
	require.NoError(t, err)
	require.Len(t, pending, 1, "and sends one link")
	assert.Equal(t, TypeAccountEmail, pending[0].Type)
	assert.Equal(t, accountEmailMaxRetry, pending[0].MaxRetry)
	assert.Equal(t, accountEmailRetention, pending[0].Retention)
	assert.NotContains(t, string(pending[0].Payload), "token", "the task never holds a link")

	decoded, err := decodeAccountEmail(asynq.NewTask(pending[0].Type, pending[0].Payload))
	require.NoError(t, err)
	assert.Equal(t, task, decoded)
}

type recordingAccountSender struct {
	got []port.AccountEmailTask
	err error
}

func (s *recordingAccountSender) Send(_ context.Context, task port.AccountEmailTask) error {
	s.got = append(s.got, task)
	return s.err
}

func TestAccountEmailHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	payload := func(t *testing.T, p accountEmailPayload) *asynq.Task {
		t.Helper()
		raw, err := json.Marshal(p)
		require.NoError(t, err)
		return asynq.NewTask(TypeAccountEmail, raw)
	}
	valid := accountEmailPayload{EventID: uuid.NewString(), UserID: uuid.NewString(), Purpose: string(domainaccount.PurposeVerifyEmail)}

	t.Run("hands_the_task_to_the_sender", func(t *testing.T) {
		t.Parallel()
		sender := &recordingAccountSender{}
		require.NoError(t, AccountEmailHandler(sender)(ctx, payload(t, valid)))
		require.Len(t, sender.got, 1)
		assert.Equal(t, valid.UserID, sender.got[0].UserID.String())
	})

	t.Run("an_undeliverable_email_is_not_retried", func(t *testing.T) {
		t.Parallel()
		sender := &recordingAccountSender{err: port.ErrEmailUndeliverable}
		require.ErrorIs(t, AccountEmailHandler(sender)(ctx, payload(t, valid)), asynq.SkipRetry)
	})

	t.Run("a_task_it_cannot_read_is_not_retried", func(t *testing.T) {
		t.Parallel()
		sender := &recordingAccountSender{}
		for _, bad := range []*asynq.Task{
			asynq.NewTask(TypeAccountEmail, []byte("{")),
			payload(t, accountEmailPayload{EventID: valid.EventID, UserID: "x", Purpose: valid.Purpose}),
			payload(t, accountEmailPayload{EventID: valid.EventID, UserID: valid.UserID, Purpose: "login"}),
			payload(t, accountEmailPayload{EventID: "x", UserID: valid.UserID, Purpose: valid.Purpose}),
		} {
			require.ErrorIs(t, AccountEmailHandler(sender)(ctx, bad), asynq.SkipRetry, string(bad.Payload()))
		}
		assert.Empty(t, sender.got)
	})
}
