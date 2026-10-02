package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	domainaccount "github.com/boms/backend/internal/domain/account"
	"github.com/boms/backend/internal/port"
)

// TypeAccountEmail is the task that emails a confirmation or reset link.
const TypeAccountEmail = "email:account"

const (
	// Five retries span about ten minutes. Someone waiting for a link asks
	// again sooner than that; a later one would only be a surprise.
	accountEmailMaxRetry = 5
	// Long enough for an event delivered again to find the task and send no
	// second link; redelivery happens within minutes.
	accountEmailRetention = time.Hour
)

type accountEmailPayload struct {
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
	Purpose string `json:"purpose"`
}

// EnqueueAccountEmail implements port.AccountEmailQueue; the task id is the
// event id, and the task holds ids only — the link is made when it is sent.
func (q *EmailQueue) EnqueueAccountEmail(ctx context.Context, task port.AccountEmailTask) error {
	payload, err := json.Marshal(accountEmailPayload{
		EventID: task.EventID.String(),
		UserID:  task.UserID.String(),
		Purpose: string(task.Purpose),
	})
	if err != nil {
		return fmt.Errorf("encode account email task: %w", err)
	}
	return q.enqueue(ctx, asynq.NewTask(TypeAccountEmail, payload), QueueEmail, task.EventID.String(), accountEmailMaxRetry, accountEmailRetention)
}

// accountEmailSender is the worker side of an account email.
type accountEmailSender interface {
	Send(ctx context.Context, task port.AccountEmailTask) error
}

// AccountEmailHandler runs account email tasks, like OrderEmailHandler.
func AccountEmailHandler(sender accountEmailSender) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		task, err := decodeAccountEmail(t)
		if err != nil {
			return fmt.Errorf("%w: %w", err, asynq.SkipRetry)
		}
		return runEmail(sender.Send(ctx, task), fmt.Sprintf("user %s %s", task.UserID, task.Purpose))
	}
}

func decodeAccountEmail(t *asynq.Task) (port.AccountEmailTask, error) {
	var payload accountEmailPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return port.AccountEmailTask{}, fmt.Errorf("decode account email task: %w", err)
	}
	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		return port.AccountEmailTask{}, fmt.Errorf("account email task user id: %w", err)
	}
	purpose := domainaccount.Purpose(payload.Purpose)
	if !purpose.Valid() {
		return port.AccountEmailTask{}, fmt.Errorf("account email task purpose %q is unknown", payload.Purpose)
	}
	eventID, err := uuid.Parse(payload.EventID)
	if err != nil {
		return port.AccountEmailTask{}, fmt.Errorf("account email task event id: %w", err)
	}
	return port.AccountEmailTask{EventID: eventID, UserID: userID, Purpose: purpose}, nil
}

var _ port.AccountEmailQueue = (*EmailQueue)(nil)
