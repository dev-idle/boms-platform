package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

type recordingAccountQueue struct {
	tasks    []port.AccountEmailTask
	failWith error
}

func (q *recordingAccountQueue) EnqueueAccountEmail(_ context.Context, task port.AccountEmailTask) error {
	if q.failWith != nil {
		return q.failWith
	}
	q.tasks = append(q.tasks, task)
	return nil
}

func TestAccountEmails_Publish(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := uuid.New()

	t.Run("queues_the_link_each_request_asks_for", func(t *testing.T) {
		t.Parallel()
		events := []domainevent.Event{
			domainaccount.VerificationRequestedEvent(userID),
			domainorder.CreatedEvent(domainorder.Order{ID: uuid.New(), UserID: &userID}),
			domainaccount.PasswordResetRequestedEvent(userID),
		}
		queue := &recordingAccountQueue{}

		require.NoError(t, NewAccountEmails(queue).Publish(ctx, events))

		assert.Equal(t, []port.AccountEmailTask{
			{EventID: events[0].ID, UserID: userID, Purpose: domainaccount.PurposeVerifyEmail},
			{EventID: events[2].ID, UserID: userID, Purpose: domainaccount.PurposeResetPassword},
		}, queue.tasks, "order events are not this publisher's")
	})

	t.Run("a_queue_failure_fails_the_delivery", func(t *testing.T) {
		t.Parallel()
		down := errors.New("redis unavailable")
		err := NewAccountEmails(&recordingAccountQueue{failWith: down}).Publish(ctx, []domainevent.Event{
			domainaccount.VerificationRequestedEvent(userID),
		})
		require.ErrorIs(t, err, down)
	})
}
