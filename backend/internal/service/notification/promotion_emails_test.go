package notification

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domainevent "github.com/boms/backend/internal/domain/event"
	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	"github.com/boms/backend/internal/port"
)

type recordingPromotionQueue struct {
	port.PromotionEmailQueue
	got []port.PromotionTask
}

func (q *recordingPromotionQueue) EnqueuePromotion(_ context.Context, task port.PromotionTask) error {
	q.got = append(q.got, task)
	return nil
}

func TestPromotionEmails(t *testing.T) {
	t.Parallel()
	queue := &recordingPromotionQueue{}
	promotionID := uuid.New()
	created := domainpromotion.CreatedEvent(promotionID)

	err := NewPromotionEmails(queue).Publish(context.Background(), []domainevent.Event{
		created,
		domainpromotion.SentEvent(promotionID),
		domainaccount.VerificationRequestedEvent(uuid.New()),
	})

	require.NoError(t, err)
	assert.Equal(t, []port.PromotionTask{{EventID: created.ID, PromotionID: promotionID}}, queue.got,
		"only a promotion created is queued, under its event")
}
