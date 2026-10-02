// Package queue runs background work on Asynq over the Redis the rest of the
// system already uses: the API and the outbox sweeper enqueue, cmd/worker runs.
package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

const (
	// TypeOrderEmail is the task that emails an order notice.
	TypeOrderEmail = "email:order"
	// QueueEmail holds the email customers wait for — order updates and
	// account links — apart from any other work.
	QueueEmail = "email"
	// QueuePromotion holds promotions, which the worker sends only while no
	// email customers wait for is queued: a promotion to every customer never
	// holds up a reset link.
	QueuePromotion = "promotion"

	// With Asynq's backoff, eight retries span about an hour and a half; past
	// that an order email is stale, and the order page has said the same thing
	// all along.
	orderEmailMaxRetry = 8
	// A send takes up to twice the mail send timeout, which config caps at 10 s,
	// plus a few database reads: the task's deadline must outlast that, or
	// Asynq would retry a send still under way and mail the customer twice.
	emailTaskTimeout = time.Minute
	// A sent task is kept this long, so an event delivered again in that time —
	// the outbox resends one whose delivery it could not record — finds it and
	// mails nobody twice. Redelivery happens within minutes.
	orderEmailRetention = 24 * time.Hour
)

type orderEmailPayload struct {
	EventID string `json:"event_id"`
	OrderID string `json:"order_id"`
	Notice  string `json:"notice"`
}

// EmailQueue queues the emails events ask for, one task per event.
type EmailQueue struct {
	client *asynq.Client
}

// NewEmailQueue returns a queue on rdb. The caller keeps rdb and closes it.
func NewEmailQueue(rdb redis.UniversalClient) *EmailQueue {
	return &EmailQueue{client: asynq.NewClientFromRedisClient(rdb)}
}

// enqueue adds task to queue under id. A second task with the same id is
// refused while the first is queued or kept after sending, and that refusal is
// the success the email ports promise: the event it came from was delivered again.
func (q *EmailQueue) enqueue(ctx context.Context, task *asynq.Task, queue, id string, maxRetry int, retention time.Duration) error {
	_, err := q.client.EnqueueContext(ctx, task,
		asynq.TaskID(id),
		asynq.Queue(queue),
		asynq.MaxRetry(maxRetry),
		asynq.Timeout(emailTaskTimeout),
		asynq.Retention(retention),
	)
	if err != nil && !errors.Is(err, asynq.ErrTaskIDConflict) {
		return fmt.Errorf("enqueue %s: %w", task.Type(), err)
	}
	return nil
}

// runEmail turns a send's outcome into the task's: an email no retry can send
// is not retried, and Asynq archives it for someone to look at.
func runEmail(err error, subject string) error {
	if err == nil {
		return nil
	}
	// Ids only: the error is logged and kept on the task.
	err = fmt.Errorf("%s email: %w", subject, err)
	if errors.Is(err, port.ErrEmailUndeliverable) {
		return fmt.Errorf("%w: %w", err, asynq.SkipRetry)
	}
	return err
}

// EnqueueOrderEmail implements port.OrderEmailQueue; the task id is the event id.
func (q *EmailQueue) EnqueueOrderEmail(ctx context.Context, task port.OrderEmailTask) error {
	payload, err := json.Marshal(orderEmailPayload{
		EventID: task.EventID.String(),
		OrderID: task.OrderID.String(),
		Notice:  string(task.Notice),
	})
	if err != nil {
		return fmt.Errorf("encode order email task: %w", err)
	}
	return q.enqueue(ctx, asynq.NewTask(TypeOrderEmail, payload), QueueEmail, task.EventID.String(), orderEmailMaxRetry, orderEmailRetention)
}

// orderEmailSender is the worker side of an order email.
type orderEmailSender interface {
	Send(ctx context.Context, task port.OrderEmailTask) error
}

// OrderEmailHandler runs order email tasks. A task it cannot read, or an email
// no retry can send, is not retried: Asynq archives it for someone to look at.
func OrderEmailHandler(sender orderEmailSender) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		task, err := decodeOrderEmail(t)
		if err != nil {
			return fmt.Errorf("%w: %w", err, asynq.SkipRetry)
		}
		return runEmail(sender.Send(ctx, task), fmt.Sprintf("order %s %s", task.OrderID, task.Notice))
	}
}

func decodeOrderEmail(t *asynq.Task) (port.OrderEmailTask, error) {
	var payload orderEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return port.OrderEmailTask{}, fmt.Errorf("decode order email task: %w", err)
	}
	orderID, err := uuid.Parse(payload.OrderID)
	if err != nil {
		return port.OrderEmailTask{}, fmt.Errorf("order email task order id: %w", err)
	}
	notice := domainorder.Notice(payload.Notice)
	if !notice.Valid() {
		return port.OrderEmailTask{}, fmt.Errorf("order email task notice %q is unknown", payload.Notice)
	}
	eventID, err := uuid.Parse(payload.EventID)
	if err != nil {
		return port.OrderEmailTask{}, fmt.Errorf("order email task event id: %w", err)
	}
	return port.OrderEmailTask{EventID: eventID, OrderID: orderID, Notice: notice}, nil
}

var _ port.OrderEmailQueue = (*EmailQueue)(nil)
