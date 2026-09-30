// Command worker delivers the outbox events the API could not deliver right
// after commit — a crashed process, a Redis outage — deletes delivery records
// past their retention, sends the emails events queue, and expires orders not
// paid in time. Run it beside the API; several copies can run at once, since
// sweeps skip rows another copy holds, each queued email goes to one of them,
// and an order expires only from awaiting payment.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/adapter/email"
	"github.com/boms/backend/internal/adapter/paypal"
	"github.com/boms/backend/internal/adapter/queue"
	postgresrepo "github.com/boms/backend/internal/adapter/repository/postgres"
	redisrepo "github.com/boms/backend/internal/adapter/repository/redis"
	"github.com/boms/backend/internal/bootstrap"
	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/infrastructure/logger"
	"github.com/boms/backend/internal/service/eventdispatch"
	"github.com/boms/backend/internal/usecase"
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
		bootstrap.EventPublisher(redisClient.RDB()),
		pgPool,
		zlog,
		cfg.Outbox.DispatchTimeout,
	)

	// Expiring an order writes events: deliver them at once, as the API does.
	pgPool.OnCommit(dispatcher.AfterCommit)

	stopEmails, err := startEmails(cfg.Mail, cfg.App.SiteURL, redisClient.RDB(), pgPool, zlog)
	if err != nil {
		zlog.Fatal("email_init", zap.Error(err))
	}

	payments := usecase.NewPaymentUsecase(pgPool, postgresrepo.NewOrderRepository(pgPool),
		postgresrepo.NewDiscountCodeRepository(pgPool), postgresrepo.NewTicketRepository(pgPool),
		postgresrepo.NewPaymentRepository(pgPool), paypal.New(cfg.PayPal), postgresrepo.NewOutboxRepository(pgPool),
		cfg.App.SiteURL, zlog)

	zlog.Info("worker_started",
		zap.Duration("sweep_interval", cfg.Outbox.SweepInterval),
		zap.Duration("prune_interval", cfg.Outbox.PruneInterval),
		zap.Duration("order_expiry_interval", cfg.Order.ExpiryInterval),
		zap.Int("email_concurrency", cfg.Mail.Concurrency),
	)
	run(ctx, dispatcher, payments, cfg, zlog)
	// A delivery is a publish and a mark, each within the dispatch timeout.
	waitCtx, cancel := context.WithTimeout(context.Background(), 2*cfg.Outbox.DispatchTimeout)
	defer cancel()
	dispatcher.Wait(waitCtx)
	// Emails being sent finish, or go back to the queue for another worker.
	stopEmails()
	zlog.Info("worker_stopped")
}

// startEmails sends queued emails through the configured SMTP server, a few at
// a time, until the returned stop is called.
func startEmails(cfg config.MailConfig, siteURL string, rdb *redis.Client, pool *postgresrepo.Pool, log *zap.Logger) (func(), error) {
	orderComposer, err := email.NewOrderComposer(siteURL)
	if err != nil {
		return nil, err
	}
	accountComposer, err := email.NewAccountComposer(siteURL)
	if err != nil {
		return nil, err
	}
	mailer := email.NewSMTPMailer(cfg)
	users := postgresrepo.NewUserRepository(pool)
	orderEmails := usecase.NewOrderEmailUsecase(postgresrepo.NewOrderRepository(pool), orderComposer, mailer, log)
	accountEmails := usecase.NewAccountEmailUsecase(users, postgresrepo.NewUserTokenRepository(pool), accountComposer, mailer, log)
	server := asynq.NewServerFromRedisClient(rdb, asynq.Config{
		Concurrency: cfg.Concurrency,
		Queues:      map[string]int{queue.QueueEmail: 1},
		// zap's sugared logger has the methods Asynq logs through.
		Logger: log.Sugar(),
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			eventID, _ := asynq.GetTaskID(ctx)
			retried, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			fields := []zap.Field{
				zap.String("type", task.Type()),
				zap.String("event_id", eventID),
				zap.Int("retried", retried),
				zap.Int("max_retry", maxRetry),
				zap.Error(err),
			}
			if errors.Is(err, asynq.SkipRetry) || retried >= maxRetry {
				// Archived: nobody gets this email unless someone looks.
				log.Error("email_task_given_up", fields...)
				return
			}
			log.Warn("email_task_failed", fields...)
		}),
		// A send already under way takes up to twice the send timeout.
		ShutdownTimeout: 2*cfg.SendTimeout + 5*time.Second,
	})
	mux := asynq.NewServeMux()
	mux.Handle(queue.TypeOrderEmail, queue.OrderEmailHandler(orderEmails))
	mux.Handle(queue.TypeAccountEmail, queue.AccountEmailHandler(accountEmails))
	if err := server.Start(mux); err != nil {
		return nil, fmt.Errorf("start email queue: %w", err)
	}
	return server.Shutdown, nil
}

// outboxJobs is the outbox work the worker schedules.
type outboxJobs interface {
	Sweep(ctx context.Context, grace time.Duration, batch int32) (int, error)
	Prune(ctx context.Context, retention time.Duration) (int64, error)
}

// orderJobs is the order work the worker schedules.
type orderJobs interface {
	ExpireOverdue(ctx context.Context) (int, error)
}

// run sweeps, prunes and expires on their intervals until ctx is cancelled.
func run(ctx context.Context, jobs outboxJobs, orders orderJobs, cfg *config.Config, log *zap.Logger) {
	// A restarted worker recovers at once instead of one interval later.
	sweep(ctx, jobs, cfg.Outbox, log)

	sweepTicker := time.NewTicker(cfg.Outbox.SweepInterval)
	defer sweepTicker.Stop()
	pruneTicker := time.NewTicker(cfg.Outbox.PruneInterval)
	defer pruneTicker.Stop()
	expiryTicker := time.NewTicker(cfg.Order.ExpiryInterval)
	defer expiryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sweepTicker.C:
			sweep(ctx, jobs, cfg.Outbox, log)
		case <-expiryTicker.C:
			// A failed order is overdue still, so the next tick tries it again.
			expired, err := orders.ExpireOverdue(ctx)
			if err != nil {
				log.Error("order_expiry_failed", zap.Error(err))
			}
			if expired > 0 {
				log.Info("orders_expired", zap.Int("orders", expired))
			}
		case <-pruneTicker.C:
			deleted, err := jobs.Prune(ctx, cfg.Outbox.Retention)
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
