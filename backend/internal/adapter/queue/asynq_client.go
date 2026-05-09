package queue

import (
	"fmt"

	"github.com/boms/backend/internal/config"
	"github.com/hibiken/asynq"
)

// NewAsynqClient returns an Asynq client for enqueueing tasks. Workers are not started here.
func NewAsynqClient(cfg config.RedisConfig) (*asynq.Client, error) {
	opts := asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	client := asynq.NewClient(opts)
	if client == nil {
		return nil, fmt.Errorf("asynq client is nil")
	}
	return client, nil
}
