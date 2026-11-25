package queue

import (
	"context"
	"fmt"
	"slices"

	"github.com/hibiken/asynq"
	goredis "github.com/redis/go-redis/v9"

	"github.com/boms/backend/internal/port"
)

// EmailFailures counts the emails the worker gave up on: the tasks archived
// after their last retry, in the email and the promotion queues. Asynq keeps an
// archived task for up to 90 days.
type EmailFailures struct {
	inspector *asynq.Inspector
}

func NewEmailFailures(rdb *goredis.Client) *EmailFailures {
	return &EmailFailures{inspector: asynq.NewInspectorFromRedisClient(rdb)}
}

// Count implements port.EmailFailures.
func (f *EmailFailures) Count(context.Context) (int64, error) {
	queues, err := f.inspector.Queues()
	if err != nil {
		return 0, fmt.Errorf("list queues: %w", err)
	}
	var total int64
	for _, name := range []string{QueueEmail, QueuePromotion} {
		// A queue no email has gone through yet has failed none.
		if !slices.Contains(queues, name) {
			continue
		}
		info, err := f.inspector.GetQueueInfo(name)
		if err != nil {
			return 0, fmt.Errorf("read %s queue: %w", name, err)
		}
		total += int64(info.Archived)
	}
	return total, nil
}

var _ port.EmailFailures = (*EmailFailures)(nil)
