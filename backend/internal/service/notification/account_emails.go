package notification

import (
	"context"
	"fmt"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domainevent "github.com/boms/backend/internal/domain/event"
	"github.com/boms/backend/internal/port"
)

// AccountEmails is an event publisher that queues the confirmation and reset
// emails account events ask for, as OrderEmails does for orders.
type AccountEmails struct {
	queue port.AccountEmailQueue
}

func NewAccountEmails(queue port.AccountEmailQueue) *AccountEmails {
	return &AccountEmails{queue: queue}
}

// Publish queues the emails events ask for and ignores every other event.
func (a *AccountEmails) Publish(ctx context.Context, events []domainevent.Event) error {
	for _, e := range events {
		userID, purpose, ok := domainaccount.EmailFor(e)
		if !ok {
			continue
		}
		task := port.AccountEmailTask{EventID: e.ID, UserID: userID, Purpose: purpose}
		if err := a.queue.EnqueueAccountEmail(ctx, task); err != nil {
			return fmt.Errorf("queue %s email for user %s: %w", purpose, userID, err)
		}
	}
	return nil
}

var _ port.EventPublisher = (*AccountEmails)(nil)
