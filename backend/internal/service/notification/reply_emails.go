package notification

import (
	"context"
	"fmt"

	domainconversation "github.com/boms/backend/internal/domain/conversation"
	domainevent "github.com/boms/backend/internal/domain/event"
	"github.com/boms/backend/internal/port"
)

// ReplyEmails is an event publisher that queues the email about replies a
// customer left unread, as OrderEmails does for orders.
type ReplyEmails struct {
	queue port.ReplyEmailQueue
}

func NewReplyEmails(queue port.ReplyEmailQueue) *ReplyEmails {
	return &ReplyEmails{queue: queue}
}

// Publish queues the emails events ask for and ignores every other event.
func (r *ReplyEmails) Publish(ctx context.Context, events []domainevent.Event) error {
	for _, e := range events {
		conversationID, ok := domainconversation.ReplyEmailFor(e)
		if !ok {
			continue
		}
		task := port.ReplyEmailTask{EventID: e.ID, ConversationID: conversationID}
		if err := r.queue.EnqueueReplyEmail(ctx, task); err != nil {
			return fmt.Errorf("queue reply email for conversation %s: %w", conversationID, err)
		}
	}
	return nil
}

var _ port.EventPublisher = (*ReplyEmails)(nil)
