package postgres

import (
	"context"
	"time"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
)

// AdminDashboardRepository reads the figures of the admin home.
type AdminDashboardRepository struct {
	queries *sqlcgen.Queries
}

func NewAdminDashboardRepository(pool *Pool) *AdminDashboardRepository {
	return &AdminDashboardRepository{queries: pool.Queries()}
}

var _ port.AdminDashboardRepository = (*AdminDashboardRepository)(nil)

// CountActiveUsersByRole implements port.AdminDashboardRepository.
func (r *AdminDashboardRepository) CountActiveUsersByRole(ctx context.Context) (map[domainuser.Role]int64, error) {
	rows, err := r.queries.CountActiveUsersByRole(ctx)
	if err != nil {
		return nil, mapRepoError(err, "count users by role")
	}
	out := make(map[domainuser.Role]int64, len(rows))
	for _, row := range rows {
		out[domainuser.Role(row.Role)] = row.Count
	}
	return out, nil
}

// Counts implements port.AdminDashboardRepository.
func (r *AdminDashboardRepository) Counts(
	ctx context.Context,
	weekStart time.Time,
	undeliveredAfter, stuckAfter time.Duration,
) (port.AdminDashboardCounts, error) {
	row, err := r.queries.AdminDashboardCounts(ctx, sqlcgen.AdminDashboardCountsParams{
		WeekStart:               weekStart,
		UndeliveredAfterSeconds: undeliveredAfter.Seconds(),
		StuckAfterSeconds:       stuckAfter.Seconds(),
	})
	if err != nil {
		return port.AdminDashboardCounts{}, mapRepoError(err, "count admin dashboard")
	}
	return port.AdminDashboardCounts{
		Disabled:          row.Disabled,
		NewCustomers:      row.NewCustomers,
		UnconfirmedEmails: row.UnconfirmedEmails,
		UndeliveredEvents: row.UndeliveredEvents,
		StuckRefunds:      row.StuckRefunds,
	}, nil
}
