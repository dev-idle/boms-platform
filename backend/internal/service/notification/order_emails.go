// Package notification turns committed events into the emails they call for.
package notification

import (
	"context"
	"fmt"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
)

// OrderEmails is an event publisher that queues the email each order notice
// calls for. The outbox dispatcher hands it every event beside the realtime
// bus, so an email is queued exactly when its event is delivered; an event
// delivered twice queues it once, as the queue keeps one task per event.
// Sending is at least once: see adapter/queue.
type OrderEmails struct {
	queue port.OrderEmailQueue
}

func NewOrderEmails(queue port.OrderEmailQueue) *OrderEmails {
	return &OrderEmails{queue: queue}
}

// Publish queues the emails events call for and ignores every other event.
func (o *OrderEmails) Publish(ctx context.Context, events []domainevent.Event) error {
	for _, e := range events {
		orderID, notice, ok := domainorder.NoticeFor(e)
		if !ok {
			continue
		}
		task := port.OrderEmailTask{EventID: e.ID, OrderID: orderID, Notice: notice}
		if err := o.queue.EnqueueOrderEmail(ctx, task); err != nil {
			return fmt.Errorf("queue %s email for order %s: %w", notice, orderID, err)
		}
	}
	return nil
}

var _ port.EventPublisher = (*OrderEmails)(nil)
