package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

type recordingQueue struct {
	tasks    []port.OrderEmailTask
	failWith error
}

func (q *recordingQueue) EnqueueOrderEmail(_ context.Context, task port.OrderEmailTask) error {
	if q.failWith != nil {
		return q.failWith
	}
	q.tasks = append(q.tasks, task)
	return nil
}

func TestOrderEmails_Publish(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	order := func(status domainorder.Status) domainorder.Order {
		return domainorder.Order{ID: uuid.New(), UserID: uuid.New(), Status: status}
	}

	t.Run("queues_one_task_per_notice_keyed_by_its_event", func(t *testing.T) {
		t.Parallel()
		placed, ready, confirmed := order(domainorder.StatusPending), order(domainorder.StatusReady), order(domainorder.StatusConfirmed)
		events := []domainevent.Event{
			domainorder.CreatedEvent(placed),
			domainorder.StatusChangedEvent(domainorder.StatusPending, confirmed),
			domainorder.StatusChangedEvent(domainorder.StatusInProduction, ready),
		}
		queue := &recordingQueue{}

		require.NoError(t, NewOrderEmails(queue).Publish(ctx, events))

		assert.Equal(t, []port.OrderEmailTask{
			{EventID: events[0].ID, OrderID: placed.ID, Notice: domainorder.NoticePlaced},
			{EventID: events[2].ID, OrderID: ready.ID, Notice: domainorder.NoticeReady},
		}, queue.tasks, "confirming an order sends nothing")
	})

	t.Run("a_queue_failure_fails_the_delivery", func(t *testing.T) {
		t.Parallel()
		down := errors.New("redis unavailable")
		queue := &recordingQueue{failWith: down}

		err := NewOrderEmails(queue).Publish(ctx, []domainevent.Event{domainorder.CreatedEvent(order(domainorder.StatusPending))})

		require.ErrorIs(t, err, down, "the outbox keeps the event and delivers it again")
	})
}
