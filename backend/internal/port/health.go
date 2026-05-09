package port

import "context"

// HealthResource is implemented by infrastructure dependencies checked at readiness.
type HealthResource interface {
	Name() string
	Ping(ctx context.Context) error
}
