package port

import (
	"context"
	"time"
)

// QuotaLimit is how many times something may happen per key in a window.
type QuotaLimit struct {
	Max    int
	Window time.Duration
}

// Quota lets something happen a limited number of times per key.
type Quota interface {
	// Take uses one of key's uses and reports false when limit.Max were
	// already used in the last limit.Window.
	Take(ctx context.Context, key string, limit QuotaLimit) (bool, error)
}
