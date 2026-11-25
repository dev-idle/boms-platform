package port

import (
	"context"
	"time"

	domainuser "github.com/boms/backend/internal/domain/user"
)

// AdminDashboardCounts is what the admin home counts besides the roles.
type AdminDashboardCounts struct {
	Disabled          int64
	NewCustomers      int64
	UnconfirmedEmails int64
	UndeliveredEvents int64
	StuckRefunds      int64
}

// AdminDashboardRepository reads the figures of the admin home.
type AdminDashboardRepository interface {
	// CountActiveUsersByRole counts the accounts in use; a role with none is absent.
	CountActiveUsersByRole(ctx context.Context) (map[domainuser.Role]int64, error)
	// Counts counts customers who signed up since weekStart, events held past
	// undeliveredAfter and refunds asked for past stuckAfter, with the
	// disabled accounts and the unconfirmed emails.
	Counts(ctx context.Context, weekStart time.Time, undeliveredAfter, stuckAfter time.Duration) (AdminDashboardCounts, error)
}

// WorkerHeartbeat records when the worker last ran its jobs, for the admin to
// see whether it runs.
type WorkerHeartbeat interface {
	Beat(ctx context.Context, at time.Time) error
	// LastBeat is nil when the worker has not run within the record's life.
	LastBeat(ctx context.Context) (*time.Time, error)
}

// EmailFailures counts the emails the worker gave up on after their last try.
type EmailFailures interface {
	Count(ctx context.Context) (int64, error)
}
