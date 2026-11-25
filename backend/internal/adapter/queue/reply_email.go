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

// TypeReplyEmail is the task that tells a customer the counter replied.
const TypeReplyEmail = "email:reply"

const (
	// Eight retries span about an hour and a half, as for an order email.
	replyEmailMaxRetry = 8
	// Long enough for an event delivered again to find the task and email
	// nobody twice; redelivery happens within minutes.
	replyEmailRetention = 24 * time.Hour
)

type replyEmailPayload struct {
	EventID        string `json:"event_id"`
	ConversationID string `json:"conversation_id"`
}

// EnqueueReplyEmail implements port.ReplyEmailQueue; the task id is the event
// id, and the task holds ids only.
func (q *EmailQueue) EnqueueReplyEmail(ctx context.Context, task port.ReplyEmailTask) error {
	payload, err := json.Marshal(replyEmailPayload{
		EventID:        task.EventID.String(),
		ConversationID: task.ConversationID.String(),
	})
	if err != nil {
		return fmt.Errorf("encode reply email task: %w", err)
	}
	return q.enqueue(ctx, asynq.NewTask(TypeReplyEmail, payload), QueueEmail, task.EventID.String(), replyEmailMaxRetry, replyEmailRetention)
}

// replyEmailSender is the worker side of a reply email.
type replyEmailSender interface {
	Send(ctx context.Context, task port.ReplyEmailTask) error
}

// ReplyEmailHandler runs reply email tasks, like OrderEmailHandler.
func ReplyEmailHandler(sender replyEmailSender) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		task, err := decodeReplyEmail(t)
		if err != nil {
			return fmt.Errorf("%w: %w", err, asynq.SkipRetry)
		}
		return runEmail(sender.Send(ctx, task), "conversation "+task.ConversationID.String())
	}
}

func decodeReplyEmail(t *asynq.Task) (port.ReplyEmailTask, error) {
	var payload replyEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return port.ReplyEmailTask{}, fmt.Errorf("decode reply email task: %w", err)
	}
	conversationID, err := uuid.Parse(payload.ConversationID)
	if err != nil {
		return port.ReplyEmailTask{}, fmt.Errorf("reply email task conversation id: %w", err)
	}
	eventID, err := uuid.Parse(payload.EventID)
	if err != nil {
		return port.ReplyEmailTask{}, fmt.Errorf("reply email task event id: %w", err)
	}
	return port.ReplyEmailTask{EventID: eventID, ConversationID: conversationID}, nil
}

var _ port.ReplyEmailQueue = (*EmailQueue)(nil)
