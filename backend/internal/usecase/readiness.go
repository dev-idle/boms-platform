package usecase

import (
	"context"
	"time"

	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
)

// Readiness aggregates infrastructure health checks for orchestration probes.
type Readiness struct {
	resources []port.HealthResource
	timeout   time.Duration
}

// NewReadiness constructs a Readiness use case with an upper bound for the full probe.
func NewReadiness(resources []port.HealthResource, overallTimeout time.Duration) *Readiness {
	return &Readiness{resources: resources, timeout: overallTimeout}
}

// Execute pings resources sequentially against one deadline-bound context.
// Sequential avoids select races between channel results and ctx.Done() under tight timeouts.
func (r *Readiness) Execute(ctx context.Context) dto.ReadinessResponse {
	timeout := r.timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	checks := make([]dto.CheckOutcome, 0, len(r.resources))
	allOK := true

	for _, res := range r.resources {
		if res == nil {
			continue
		}
		err := res.Ping(ctx)
		o := dto.CheckOutcome{Name: res.Name(), OK: err == nil}
		if err != nil {
			o.Error = err.Error()
			allOK = false
		}
		checks = append(checks, o)
	}

	status := "not_ready"
	if allOK {
		status = "ready"
	}
	return dto.ReadinessResponse{Status: status, Checks: checks}
}
