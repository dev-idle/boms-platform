package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/boms/backend/internal/port"
)

const (
	// TypePromotion is the task that queues a promotion's email for every
	// customer it goes to.
	TypePromotion = "email:promotion"
	// TypePromotionEmail is the task that emails one customer a promotion.
	TypePromotionEmail = "email:promotion_customer"

	// A promotion is not urgent: eight retries span about an hour and a half,
	// as for order emails.
	promotionMaxRetry = 8
	// Long enough for a promotion queued again — its event delivered again, or
	// its queuing retried — to find each customer's email and send none twice.
	promotionRetention = 24 * time.Hour
)

type promotionPayload struct {
	EventID     string `json:"event_id"`
	PromotionID string `json:"promotion_id"`
}

type promotionEmailPayload struct {
	PromotionID string `json:"promotion_id"`
	UserID      string `json:"user_id"`
}

// EnqueuePromotion implements port.PromotionEmailQueue; the task id is the event id.
func (q *EmailQueue) EnqueuePromotion(ctx context.Context, task port.PromotionTask) error {
	payload, err := json.Marshal(promotionPayload{EventID: task.EventID.String(), PromotionID: task.PromotionID.String()})
	if err != nil {
		return fmt.Errorf("encode promotion task: %w", err)
	}
	return q.enqueue(ctx, asynq.NewTask(TypePromotion, payload), QueuePromotion, task.EventID.String(), promotionMaxRetry, promotionRetention)
}

// EnqueuePromotionEmail implements port.PromotionEmailQueue; the task id names
// the promotion and the customer, so a customer gets each promotion once.
func (q *EmailQueue) EnqueuePromotionEmail(ctx context.Context, task port.PromotionEmailTask) error {
	payload, err := json.Marshal(promotionEmailPayload{PromotionID: task.PromotionID.String(), UserID: task.UserID.String()})
	if err != nil {
		return fmt.Errorf("encode promotion email task: %w", err)
	}
	id := "promotion:" + task.PromotionID.String() + ":" + task.UserID.String()
	return q.enqueue(ctx, asynq.NewTask(TypePromotionEmail, payload), QueuePromotion, id, promotionMaxRetry, promotionRetention)
}

// promotionSender is the worker side of a promotion.
type promotionSender interface {
	Queue(ctx context.Context, task port.PromotionTask) error
	Send(ctx context.Context, task port.PromotionEmailTask) error
}

// PromotionHandler runs promotion tasks: it queues the customers' emails.
func PromotionHandler(sender promotionSender) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload promotionPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("decode promotion task: %w: %w", err, asynq.SkipRetry)
		}
		eventID, eventErr := uuid.Parse(payload.EventID)
		promotionID, promotionErr := uuid.Parse(payload.PromotionID)
		if eventErr != nil || promotionErr != nil {
			return fmt.Errorf("promotion task ids are not valid: %w", asynq.SkipRetry)
		}
		return sender.Queue(ctx, port.PromotionTask{EventID: eventID, PromotionID: promotionID})
	}
}

// PromotionEmailHandler runs promotion email tasks, like AccountEmailHandler.
func PromotionEmailHandler(sender promotionSender) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload promotionEmailPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("decode promotion email task: %w: %w", err, asynq.SkipRetry)
		}
		promotionID, promotionErr := uuid.Parse(payload.PromotionID)
		userID, userErr := uuid.Parse(payload.UserID)
		if promotionErr != nil || userErr != nil {
			return fmt.Errorf("promotion email task ids are not valid: %w", asynq.SkipRetry)
		}
		task := port.PromotionEmailTask{PromotionID: promotionID, UserID: userID}
		return runEmail(sender.Send(ctx, task), fmt.Sprintf("promotion %s user %s", promotionID, userID))
	}
}

var _ port.PromotionEmailQueue = (*EmailQueue)(nil)
