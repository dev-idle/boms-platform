package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// pagedPromotions serves recipients after afterID, in their order in the list,
// and records what the worker asked for; any other call panics.
type pagedPromotions struct {
	port.PromotionRepository
	status     domainpromotion.Status
	recipients []uuid.UUID
	afters     []*uuid.UUID
	marked     []int32
	stillSends bool
	delivery   *port.PromotionDelivery
}

func (p *pagedPromotions) GetStatus(context.Context, uuid.UUID) (domainpromotion.Status, error) {
	return p.status, nil
}

func (p *pagedPromotions) ListRecipients(_ context.Context, afterID *uuid.UUID, limit int32) ([]uuid.UUID, error) {
	p.afters = append(p.afters, afterID)
	start := 0
	if afterID != nil {
		for i, id := range p.recipients {
			if id == *afterID {
				start = i + 1
			}
		}
	}
	end := min(start+int(limit), len(p.recipients))
	return p.recipients[start:end], nil
}

func (p *pagedPromotions) MarkSent(_ context.Context, _ uuid.UUID, recipients int32) (bool, error) {
	p.marked = append(p.marked, recipients)
	return p.stillSends, nil
}

func (p *pagedPromotions) GetDelivery(context.Context, uuid.UUID, uuid.UUID) (*port.PromotionDelivery, error) {
	if p.delivery == nil {
		return nil, apperrors.ErrNotFound
	}
	return p.delivery, nil
}

type recordingPromotionQueue struct {
	queued []port.PromotionEmailTask
	err    error
}

func (q *recordingPromotionQueue) EnqueuePromotion(context.Context, port.PromotionTask) error {
	return nil
}

func (q *recordingPromotionQueue) EnqueuePromotionEmail(_ context.Context, task port.PromotionEmailTask) error {
	if q.err != nil {
		return q.err
	}
	q.queued = append(q.queued, task)
	return nil
}

type failingPromotionComposer struct{}

func (failingPromotionComposer) ComposePromotionEmail(port.PromotionEmail) (port.Email, error) {
	return port.Email{}, errors.New("template broke")
}

type countingMailer struct{ sent int }

func (m *countingMailer) Send(context.Context, port.Email) error {
	m.sent++
	return nil
}

func newIDs(n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = uuid.New()
	}
	return ids
}

func TestPromotionEmailUsecase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tokens := domainpromotion.NewUnsubscribeTokens(strings.Repeat("u", 32))
	task := port.PromotionTask{EventID: uuid.New(), PromotionID: uuid.New()}
	newUC := func(promotions *pagedPromotions, queue *recordingPromotionQueue, events *recordingOutbox, composer port.PromotionEmailComposer, mailer port.Mailer) *usecase.PromotionEmailUsecase {
		return usecase.NewPromotionEmailUsecase(promotions, queue, passthroughTxManager{}, events, composer, mailer, tokens, zap.NewNop())
	}

	t.Run("queues_every_recipient_a_batch_at_a_time", func(t *testing.T) {
		t.Parallel()
		promotions := &pagedPromotions{status: domainpromotion.StatusSending, recipients: newIDs(103), stillSends: true}
		queue := &recordingPromotionQueue{}
		events := &recordingOutbox{}

		require.NoError(t, newUC(promotions, queue, events, nil, nil).Queue(ctx, task))

		require.Len(t, promotions.afters, 2, "a full batch, then the rest")
		assert.Nil(t, promotions.afters[0])
		assert.Equal(t, promotions.recipients[99], *promotions.afters[1], "the next batch starts after the last one read")
		assert.Len(t, queue.queued, 103)
		assert.Equal(t, []int32{103}, promotions.marked)
		require.Len(t, events.events, 1)
		assert.Equal(t, domainpromotion.TopicSent, events.events[0].Topic)
	})

	t.Run("a_full_last_batch_reads_once_more", func(t *testing.T) {
		t.Parallel()
		promotions := &pagedPromotions{status: domainpromotion.StatusSending, recipients: newIDs(100), stillSends: true}

		require.NoError(t, newUC(promotions, &recordingPromotionQueue{}, &recordingOutbox{}, nil, nil).Queue(ctx, task))

		assert.Len(t, promotions.afters, 2)
		assert.Equal(t, []int32{100}, promotions.marked)
	})

	t.Run("a_promotion_marked_sent_meanwhile_tells_nobody_twice", func(t *testing.T) {
		t.Parallel()
		promotions := &pagedPromotions{status: domainpromotion.StatusSending, recipients: newIDs(2)}
		events := &recordingOutbox{}

		require.NoError(t, newUC(promotions, &recordingPromotionQueue{}, events, nil, nil).Queue(ctx, task))

		assert.Empty(t, events.events)
	})

	t.Run("a_promotion_sent_is_not_queued_again", func(t *testing.T) {
		t.Parallel()
		promotions := &pagedPromotions{status: domainpromotion.StatusSent, recipients: newIDs(2)}
		queue := &recordingPromotionQueue{}

		require.NoError(t, newUC(promotions, queue, &recordingOutbox{}, nil, nil).Queue(ctx, task))

		assert.Empty(t, promotions.afters)
		assert.Empty(t, queue.queued)
	})

	t.Run("a_queue_that_fails_fails_the_task_for_a_retry", func(t *testing.T) {
		t.Parallel()
		promotions := &pagedPromotions{status: domainpromotion.StatusSending, recipients: newIDs(2), stillSends: true}
		queue := &recordingPromotionQueue{err: errors.New("redis down")}

		err := newUC(promotions, queue, &recordingOutbox{}, nil, nil).Queue(ctx, task)

		require.Error(t, err)
		assert.Empty(t, promotions.marked, "not marked sent before every email is queued")
	})

	t.Run("an_email_that_cannot_be_written_is_not_retried", func(t *testing.T) {
		t.Parallel()
		promotions := &pagedPromotions{delivery: &port.PromotionDelivery{Subject: "S", Body: "B", Email: "mai@example.com"}}
		mailer := &countingMailer{}

		err := newUC(promotions, nil, nil, failingPromotionComposer{}, mailer).
			Send(ctx, port.PromotionEmailTask{PromotionID: task.PromotionID, UserID: uuid.New()})

		require.ErrorIs(t, err, port.ErrEmailUndeliverable)
		assert.Zero(t, mailer.sent)
	})

	t.Run("a_customer_no_longer_agreeing_gets_nothing", func(t *testing.T) {
		t.Parallel()
		mailer := &countingMailer{}

		err := newUC(&pagedPromotions{}, nil, nil, failingPromotionComposer{}, mailer).
			Send(ctx, port.PromotionEmailTask{PromotionID: task.PromotionID, UserID: uuid.New()})

		require.NoError(t, err)
		assert.Zero(t, mailer.sent)
	})
}
