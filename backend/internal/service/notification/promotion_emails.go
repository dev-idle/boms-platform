package notification

import (
	"context"
	"fmt"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	"github.com/boms/backend/internal/port"
)

// PromotionEmails is an event publisher that queues a promotion a manager
// sent, as OrderEmails does for orders.
type PromotionEmails struct {
	queue port.PromotionEmailQueue
}

func NewPromotionEmails(queue port.PromotionEmailQueue) *PromotionEmails {
	return &PromotionEmails{queue: queue}
}

// Publish queues the promotions events ask for and ignores every other event.
func (p *PromotionEmails) Publish(ctx context.Context, events []domainevent.Event) error {
	for _, e := range events {
		promotionID, ok := domainpromotion.ToSend(e)
		if !ok {
			continue
		}
		if err := p.queue.EnqueuePromotion(ctx, port.PromotionTask{EventID: e.ID, PromotionID: promotionID}); err != nil {
			return fmt.Errorf("queue promotion %s: %w", promotionID, err)
		}
	}
	return nil
}

var _ port.EventPublisher = (*PromotionEmails)(nil)
