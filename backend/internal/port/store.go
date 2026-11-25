package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainstore "github.com/boms/backend/internal/domain/store"
)

// HeldPickupRules is what a booking must still meet under the held settings.
type HeldPickupRules struct {
	OpensAt, ClosesAt time.Duration
	DayClosed         bool
}

// StoreSettingsRepository keeps the bakery's pickup settings and closed dates.
type StoreSettingsRepository interface {
	GetSettings(ctx context.Context) (domainstore.Settings, error)
	// GetSettingsForUpdate locks the settings until the caller's transaction ends.
	GetSettingsForUpdate(ctx context.Context) (domainstore.Settings, error)
	// HoldPickupRules holds the settings against a change until the caller's
	// transaction ends, and reads the opening hours and whether day is closed.
	HoldPickupRules(ctx context.Context, day time.Time) (HeldPickupRules, error)
	UpdateSettings(ctx context.Context, s domainstore.Settings) (domainstore.Settings, error)
	// ListClosedDates returns the closures from one day to another, inclusive,
	// earliest first.
	ListClosedDates(ctx context.Context, from, to time.Time) ([]domainstore.ClosedDate, error)
	// AddClosedDate returns apperrors.ErrConflict when the day is already closed.
	AddClosedDate(ctx context.Context, day time.Time, reason string) (domainstore.ClosedDate, error)
	// RemoveClosedDate returns apperrors.ErrNotFound when there is no such closure.
	RemoveClosedDate(ctx context.Context, id uuid.UUID) (domainstore.ClosedDate, error)
}
