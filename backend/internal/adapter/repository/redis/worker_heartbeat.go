package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/boms/backend/internal/port"
)

const (
	workerHeartbeatKey = "worker:last_run"
	// workerHeartbeatTTL outlives any job interval: once the key is gone the
	// worker has not run for that long.
	workerHeartbeatTTL = 7 * 24 * time.Hour
)

// WorkerHeartbeat keeps when the worker last ran its jobs.
type WorkerHeartbeat struct {
	rdb *redis.Client
}

func NewWorkerHeartbeat(client *Client) *WorkerHeartbeat {
	return &WorkerHeartbeat{rdb: client.RDB()}
}

// Beat implements port.WorkerHeartbeat.
func (h *WorkerHeartbeat) Beat(ctx context.Context, at time.Time) error {
	if err := h.rdb.Set(ctx, workerHeartbeatKey, at.UTC().Format(time.RFC3339Nano), workerHeartbeatTTL).Err(); err != nil {
		return fmt.Errorf("record worker run: %w", err)
	}
	return nil
}

// LastBeat implements port.WorkerHeartbeat.
func (h *WorkerHeartbeat) LastBeat(ctx context.Context) (*time.Time, error) {
	raw, err := h.rdb.Get(ctx, workerHeartbeatKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read worker run: %w", err)
	}
	at, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return nil, fmt.Errorf("read worker run: %w", err)
	}
	return &at, nil
}

var _ port.WorkerHeartbeat = (*WorkerHeartbeat)(nil)
