package usecase

import (
	"context"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
)

const (
	// undeliveredAfter is how long an event may wait in the outbox before it
	// counts as undelivered: the sweeper delivers it well within that.
	undeliveredAfter = 5 * time.Minute
	// refundStuckAfter is how long a refund may stay asked for before it
	// counts as stuck: the worker tries it on every run.
	refundStuckAfter = time.Hour
	// workerLateRuns is how many job intervals without a run make the worker late.
	workerLateRuns = 3
)

// dashboardRoles is the order the admin home lists the accounts in.
var dashboardRoles = []domainuser.Role{
	domainuser.RoleCustomer, domainuser.RoleStaff, domainuser.RoleBaker, domainuser.RoleManager, domainuser.RoleAdmin,
}

// AdminDashboardUsecase reads the admin home: the accounts, and whether the
// system runs as it should.
type AdminDashboardUsecase struct {
	dashboard   port.AdminDashboardRepository
	readiness   *Readiness
	heartbeat   port.WorkerHeartbeat
	emails      port.EmailFailures
	jobInterval time.Duration
	log         *zap.Logger
}

func NewAdminDashboardUsecase(
	dashboard port.AdminDashboardRepository,
	readiness *Readiness,
	heartbeat port.WorkerHeartbeat,
	emails port.EmailFailures,
	jobInterval time.Duration,
	log *zap.Logger,
) *AdminDashboardUsecase {
	return &AdminDashboardUsecase{
		dashboard: dashboard, readiness: readiness, heartbeat: heartbeat, emails: emails, jobInterval: jobInterval, log: log,
	}
}

// Get reads the figures as of now, all at once. What lives in Redis — the
// worker's last run and the email queue — is reported unknown when Redis
// cannot be read, as the health check reports Redis down; the database's
// figures are the page. No part cancels another: a database failure must not
// show up as Redis failing too.
func (u *AdminDashboardUsecase) Get(ctx context.Context, now time.Time) (*dto.AdminDashboardResponse, error) {
	var (
		roles   map[domainuser.Role]int64
		counts  port.AdminDashboardCounts
		health  dto.ReadinessResponse
		lastRun *time.Time
		failed  *int64
	)
	var group errgroup.Group
	group.Go(func() (err error) {
		roles, err = u.dashboard.CountActiveUsersByRole(ctx)
		return err
	})
	group.Go(func() (err error) {
		counts, err = u.dashboard.Counts(ctx, weekStart(now), undeliveredAfter, refundStuckAfter)
		return err
	})
	group.Go(func() error {
		health = u.readiness.Execute(ctx)
		return nil
	})
	group.Go(func() error {
		at, err := u.heartbeat.LastBeat(ctx)
		if err != nil {
			u.log.Warn("admin_dashboard_worker_unread", zap.Error(err))
			return nil
		}
		lastRun = at
		return nil
	})
	group.Go(func() error {
		n, err := u.emails.Count(ctx)
		if err != nil {
			u.log.Warn("admin_dashboard_email_queue_unread", zap.Error(err))
			return nil
		}
		failed = &n
		return nil
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}

	byRole := make([]dto.AdminRoleCount, 0, len(dashboardRoles))
	for _, role := range dashboardRoles {
		byRole = append(byRole, dto.AdminRoleCount{Role: string(role), Count: roles[role]})
	}
	return &dto.AdminDashboardResponse{
		Accounts: dto.AdminDashboardAccounts{
			ByRole:               byRole,
			Disabled:             counts.Disabled,
			NewCustomersThisWeek: counts.NewCustomers,
			UnconfirmedEmails:    counts.UnconfirmedEmails,
		},
		Health: dto.AdminDashboardHealth{
			Database:          health.Checks.DB,
			Redis:             health.Checks.Redis,
			Worker:            workerState(lastRun, now, u.jobInterval),
			WorkerLastRunAt:   lastRun,
			UndeliveredEvents: counts.UndeliveredEvents,
			FailedEmails:      failed,
			StuckRefunds:      counts.StuckRefunds,
		},
	}, nil
}

// workerState is up when the worker ran within a few job intervals, late when
// it has not, and unknown with no run on record.
func workerState(lastRun *time.Time, now time.Time, jobInterval time.Duration) string {
	switch {
	case lastRun == nil:
		return dto.WorkerUnknown
	case now.Sub(*lastRun) > workerLateRuns*jobInterval:
		return dto.WorkerLate
	default:
		return dto.WorkerUp
	}
}

// weekStart is the start of the bakery's week holding now: Monday at midnight,
// bakery time.
func weekStart(now time.Time) time.Time {
	day := domainstore.DayOf(now)
	monday := day.AddDate(0, 0, -((int(day.Weekday()) + 6) % 7))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, domainstore.Location)
}
