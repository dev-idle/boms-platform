package redis

import (
	"context"
	"fmt"

	"github.com/boms/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

// Client wraps go-redis with pooling configured from application config.
type Client struct {
	inner *redis.Client
}

// NewClient dials Redis and verifies connectivity.
func NewClient(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	pingCtx, cancel := context.WithTimeout(ctx, cfg.HealthCheckTimeout)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{inner: rdb}, nil
}

// RDB exposes the underlying client for cache, rate limiting backends, or Asynq.
func (c *Client) RDB() *redis.Client {
	return c.inner
}

// Close closes the client.
func (c *Client) Close() error {
	if c == nil || c.inner == nil {
		return nil
	}
	return c.inner.Close()
}

// Name implements port.HealthResource.
func (c *Client) Name() string {
	return "redis"
}

// Ping implements port.HealthResource.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.inner == nil {
		return fmt.Errorf("redis client is nil")
	}
	return c.inner.Ping(ctx).Err()
}
