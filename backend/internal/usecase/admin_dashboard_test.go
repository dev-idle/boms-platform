package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

type dashboardFigures struct {
	roles  map[domainuser.Role]int64
	counts port.AdminDashboardCounts
	week   time.Time
	err    error
}

func (d *dashboardFigures) CountActiveUsersByRole(context.Context) (map[domainuser.Role]int64, error) {
	return d.roles, d.err
}

func (d *dashboardFigures) Counts(_ context.Context, weekStart time.Time, _, _ time.Duration) (port.AdminDashboardCounts, error) {
	d.week = weekStart
	return d.counts, d.err
}

type lastRun struct {
	at  *time.Time
	err error
}

func (l lastRun) Beat(context.Context, time.Time) error        { return nil }
func (l lastRun) LastBeat(context.Context) (*time.Time, error) { return l.at, l.err }

type failedEmails struct {
	count int64
	err   error
}

func (f failedEmails) Count(context.Context) (int64, error) { return f.count, f.err }

func TestAdminDashboardUsecase_Get(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	interval := time.Minute
	readiness := func(redis error) *usecase.Readiness {
		return usecase.NewReadiness([]port.HealthResource{fakeResource{name: "postgres"}, fakeResource{name: "redis", err: redis}}, time.Second, zap.NewNop())
	}

	t.Run("lists_every_role_and_reports_a_worker_that_runs", func(t *testing.T) {
		t.Parallel()
		figures := &dashboardFigures{
			roles:  map[domainuser.Role]int64{domainuser.RoleCustomer: 12, domainuser.RoleAdmin: 1},
			counts: port.AdminDashboardCounts{Disabled: 2, NewCustomers: 3, UnconfirmedEmails: 4, UndeliveredEvents: 5, StuckRefunds: 6},
		}
		ran := time.Now().Add(-interval)
		uc := usecase.NewAdminDashboardUsecase(figures, readiness(nil), lastRun{at: &ran}, failedEmails{count: 7}, interval, zap.NewNop())

		got, err := uc.Get(ctx, time.Now())

		require.NoError(t, err)
		assert.Equal(t, []dto.AdminRoleCount{
			{Role: "customer", Count: 12}, {Role: "staff"}, {Role: "baker"}, {Role: "manager"}, {Role: "admin", Count: 1},
		}, got.Accounts.ByRole, "a role with no account still shows")
		assert.Equal(t, dto.AdminDashboardAccounts{
			ByRole: got.Accounts.ByRole, Disabled: 2, NewCustomersThisWeek: 3, UnconfirmedEmails: 4,
		}, got.Accounts)
		assert.Equal(t, "up", got.Health.Database)
		assert.Equal(t, "up", got.Health.Redis)
		assert.Equal(t, dto.WorkerUp, got.Health.Worker)
		require.NotNil(t, got.Health.FailedEmails)
		assert.Equal(t, int64(7), *got.Health.FailedEmails)
		assert.Equal(t, int64(5), got.Health.UndeliveredEvents)
		assert.Equal(t, int64(6), got.Health.StuckRefunds)
	})

	t.Run("counts_new_customers_from_monday_at_midnight_bakery_time", func(t *testing.T) {
		t.Parallel()
		figures := &dashboardFigures{}
		uc := usecase.NewAdminDashboardUsecase(figures, readiness(nil), lastRun{}, failedEmails{}, interval, zap.NewNop())
		// Sunday 18:00 UTC is already Monday 01:00 at the bakery.
		now := time.Date(2026, 11, 22, 18, 0, 0, 0, time.UTC)

		_, err := uc.Get(ctx, now)

		require.NoError(t, err)
		assert.True(t, figures.week.Equal(time.Date(2026, 11, 23, 0, 0, 0, 0, domainstore.Location)), "got %s", figures.week)
	})

	t.Run("a_database_failure_fails_the_page", func(t *testing.T) {
		t.Parallel()
		down := errors.New("database is down")
		uc := usecase.NewAdminDashboardUsecase(&dashboardFigures{err: down}, readiness(nil), lastRun{}, failedEmails{}, interval, zap.NewNop())

		got, err := uc.Get(ctx, time.Now())

		require.ErrorIs(t, err, down)
		assert.Nil(t, got)
	})

	t.Run("a_worker_that_stopped_is_late", func(t *testing.T) {
		t.Parallel()
		ran := time.Now().Add(-10 * interval)
		uc := usecase.NewAdminDashboardUsecase(&dashboardFigures{}, readiness(nil), lastRun{at: &ran}, failedEmails{}, interval, zap.NewNop())

		got, err := uc.Get(ctx, time.Now())

		require.NoError(t, err)
		assert.Equal(t, dto.WorkerLate, got.Health.Worker)
	})

	t.Run("what_lives_in_redis_is_unknown_while_redis_is_down", func(t *testing.T) {
		t.Parallel()
		down := errors.New("redis is down")
		uc := usecase.NewAdminDashboardUsecase(&dashboardFigures{}, readiness(down), lastRun{err: down}, failedEmails{err: down}, interval, zap.NewNop())

		got, err := uc.Get(ctx, time.Now())

		require.NoError(t, err, "the database's figures are still the page")
		assert.Equal(t, "down", got.Health.Redis)
		assert.Equal(t, dto.WorkerUnknown, got.Health.Worker)
		assert.Nil(t, got.Health.WorkerLastRunAt)
		assert.Nil(t, got.Health.FailedEmails)
	})
}
