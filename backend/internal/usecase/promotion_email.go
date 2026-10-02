package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// promotionRecipientBatch is how many customers one read of a promotion's
// recipients takes.
const promotionRecipientBatch int32 = 100

// PromotionEmailUsecase runs in the worker: it queues a promotion's email for
// every customer it goes to, a batch at a time, and sends each one. Every email
// is its own task, so one that fails is retried alone and a promotion queued
// again mails nobody twice.
type PromotionEmailUsecase struct {
	promotions port.PromotionRepository
	queue      port.PromotionEmailQueue
	tx         port.TxManager
	events     port.EventOutbox
	composer   port.PromotionEmailComposer
	mailer     port.Mailer
	tokens     domainpromotion.UnsubscribeTokens
	log        *zap.Logger
}

func NewPromotionEmailUsecase(
	promotions port.PromotionRepository,
	queue port.PromotionEmailQueue,
	tx port.TxManager,
	events port.EventOutbox,
	composer port.PromotionEmailComposer,
	mailer port.Mailer,
	tokens domainpromotion.UnsubscribeTokens,
	log *zap.Logger,
) *PromotionEmailUsecase {
	return &PromotionEmailUsecase{
		promotions: promotions,
		queue:      queue,
		tx:         tx,
		events:     events,
		composer:   composer,
		mailer:     mailer,
		tokens:     tokens,
		log:        log,
	}
}

// Queue queues the promotion's email for every customer it goes to now, then
// records how many and tells managers. A promotion already sent is left alone.
func (u *PromotionEmailUsecase) Queue(ctx context.Context, task port.PromotionTask) error {
	status, err := u.promotions.GetStatus(ctx, task.PromotionID)
	if err != nil {
		return err
	}
	if status == domainpromotion.StatusSent {
		return nil
	}
	var queued int64
	var after *uuid.UUID
	for {
		batch, err := u.promotions.ListRecipients(ctx, after, promotionRecipientBatch)
		if err != nil {
			return err
		}
		for _, userID := range batch {
			if err := u.queue.EnqueuePromotionEmail(ctx, port.PromotionEmailTask{PromotionID: task.PromotionID, UserID: userID}); err != nil {
				return fmt.Errorf("queue promotion %s for user %s: %w", task.PromotionID, userID, err)
			}
		}
		queued += int64(len(batch))
		if len(batch) < int(promotionRecipientBatch) {
			break
		}
		after = &batch[len(batch)-1]
	}
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		sent, err := u.promotions.MarkSent(txCtx, task.PromotionID, utils.Int32FromInt64(queued))
		if err != nil || !sent {
			return err
		}
		return u.events.Add(txCtx, domainpromotion.SentEvent(task.PromotionID))
	})
	if err != nil {
		return err
	}
	u.log.Info("promotion_queued", zap.String("promotion_id", task.PromotionID.String()), zap.Int64("recipients", queued))
	return nil
}

// Send emails the promotion to one customer, while promotions still go to
// them: one who withdrew their agreement, closed their account or is no longer
// a customer since it was queued gets nothing. Errors wrapping
// port.ErrEmailUndeliverable are final; any other may pass on a retry.
func (u *PromotionEmailUsecase) Send(ctx context.Context, task port.PromotionEmailTask) error {
	delivery, err := u.promotions.GetDelivery(ctx, task.PromotionID, task.UserID)
	if errors.Is(err, apperrors.ErrNotFound) {
		u.log.Info("promotion_email_skipped",
			zap.String("promotion_id", task.PromotionID.String()),
			zap.String("user_id", task.UserID.String()),
			zap.String("reason", "not_a_recipient"),
		)
		return nil
	}
	if err != nil {
		return err
	}
	var name string
	if delivery.Name != nil {
		name = *delivery.Name
	}
	email, err := u.composer.ComposePromotionEmail(port.PromotionEmail{
		To:               delivery.Email,
		ToName:           name,
		Subject:          delivery.Subject,
		Body:             delivery.Body,
		UnsubscribeToken: u.tokens.Of(task.UserID),
	})
	if err != nil {
		return fmt.Errorf("%w: compose promotion email: %w", port.ErrEmailUndeliverable, err)
	}
	if err := u.mailer.Send(ctx, email); err != nil {
		return err
	}
	u.log.Info("promotion_email_sent",
		zap.String("promotion_id", task.PromotionID.String()),
		zap.String("user_id", task.UserID.String()),
	)
	return nil
}
