// Command worker delivers the outbox events the API could not deliver right
// after commit — a crashed process, a Redis outage — and deletes delivery
// records past their retention. Run it beside the API; several copies can run
// at once, since sweeps skip rows another copy holds.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/adapter/eventbus"
	postgresrepo "github.com/boms/backend/internal/adapter/repository/postgres"
	redisrepo "github.com/boms/backend/internal/adapter/repository/redis"
	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/infrastructure/logger"
	"github.com/boms/backend/internal/service/eventdispatch"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.LoadWorker()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	zlog, err := logger.New(cfg.Log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = zlog.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pgPool, err := postgresrepo.NewPool(ctx, cfg.Postgres)
	if err != nil {
		zlog.Fatal("postgres_init", zap.Error(err))
	}
	defer pgPool.Close()

	redisClient, err := redisrepo.NewClient(ctx, cfg.Redis)
	if err != nil {
		zlog.Fatal("redis_init", zap.Error(err))
	}
	defer func() { _ = redisClient.Close() }()

	dispatcher := eventdispatch.New(
		postgresrepo.NewOutboxRepository(pgPool),
		eventbus.NewRedisPublisher(redisClient.RDB()),
		pgPool,
		zlog,
		cfg.Outbox.DispatchTimeout,
	)

	zlog.Info("worker_started",
		zap.Duration("sweep_interval", cfg.Outbox.SweepInterval),
		zap.Duration("prune_interval", cfg.Outbox.PruneInterval),
	)
	run(ctx, dispatcher, cfg.Outbox, zlog)
	zlog.Info("worker_stopped")
}

// outboxJobs is the work the worker schedules.
type outboxJobs interface {
	Sweep(ctx context.Context, grace time.Duration, batch int32) (int, error)
	Prune(ctx context.Context, retention time.Duration) (int64, error)
}

// run sweeps and prunes on their intervals until ctx is cancelled.
func run(ctx context.Context, jobs outboxJobs, cfg config.OutboxConfig, log *zap.Logger) {
	// A restarted worker recovers at once instead of one interval later.
	sweep(ctx, jobs, cfg, log)

	sweepTicker := time.NewTicker(cfg.SweepInterval)
	defer sweepTicker.Stop()
	pruneTicker := time.NewTicker(cfg.PruneInterval)
	defer pruneTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sweepTicker.C:
			sweep(ctx, jobs, cfg, log)
		case <-pruneTicker.C:
			deleted, err := jobs.Prune(ctx, cfg.Retention)
			if err != nil {
				log.Error("outbox_prune_failed", zap.Error(err))
				continue
			}
			if deleted > 0 {
				log.Info("outbox_pruned", zap.Int64("events", deleted))
			}
		}
	}
}

// sweep keeps taking batches while they come back full, so a backlog left by an
// outage drains in one tick instead of one batch per interval.
func sweep(ctx context.Context, jobs outboxJobs, cfg config.OutboxConfig, log *zap.Logger) {
	for ctx.Err() == nil {
		delivered, err := jobs.Sweep(ctx, cfg.SweepGrace, cfg.SweepBatch)
		if err != nil {
			log.Error("outbox_sweep_failed", zap.Error(err))
			return
		}
		if delivered > 0 {
			log.Info("outbox_swept", zap.Int("events", delivered))
		}
		if delivered < int(cfg.SweepBatch) {
			return
		}
	}
}
