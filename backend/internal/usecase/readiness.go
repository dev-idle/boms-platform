package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Readiness aggregates infrastructure health checks for orchestration probes.
type Readiness struct {
	resources []port.HealthResource
	timeout   time.Duration
	log       *zap.Logger
}

// NewReadiness constructs a Readiness use case with an upper bound for the full probe.
func NewReadiness(resources []port.HealthResource, overallTimeout time.Duration, log *zap.Logger) *Readiness {
	return &Readiness{resources: resources, timeout: overallTimeout, log: log}
}

// Execute pings resources in parallel. Probe errors are logged, never returned to clients.
func (r *Readiness) Execute(ctx context.Context) dto.ReadinessResponse {
	timeout := r.timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	checks := dto.ReadinessChecks{DB: "down", Redis: "down"}
	allOK := true
	var mu sync.Mutex

	g, gctx := errgroup.WithContext(ctx)
	for _, res := range r.resources {
		if res == nil {
			continue
		}
		res := res
		g.Go(func() error {
			err := res.Ping(gctx)
			ok := err == nil
			state := "down"
			if ok {
				state = "up"
			} else if r.log != nil {
				r.log.Error("readiness_check_failed",
					zap.String("resource", res.Name()),
					zap.Error(err),
				)
			}
			mu.Lock()
			if !ok {
				allOK = false
			}
			switch res.Name() {
			case "postgres":
				checks.DB = state
			case "redis":
				checks.Redis = state
			}
			mu.Unlock()
			return nil
		})
	}
	_ = g.Wait()

	status := "degraded"
	if allOK {
		status = "ok"
	}
	return dto.ReadinessResponse{Status: status, Checks: checks}
}
