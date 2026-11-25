// Command worker delivers the outbox events the API could not deliver right
// after commit — a crashed process, a Redis outage — deletes delivery records
// past their retention, sends the emails events queue, expires orders not paid
// in time, makes the refunds cancelled orders ask for and records missed
// pickups. Run it beside the API; several copies can run at once, since sweeps
// skip rows another copy holds, each queued email goes to one of them, every
// order move is guarded on the status it leaves, and a refund is recorded once.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/adapter/cloudinary"
	"github.com/boms/backend/internal/adapter/email"
	"github.com/boms/backend/internal/adapter/paypal"
	"github.com/boms/backend/internal/adapter/queue"
	postgresrepo "github.com/boms/backend/internal/adapter/repository/postgres"
	redisrepo "github.com/boms/backend/internal/adapter/repository/redis"
	"github.com/boms/backend/internal/bootstrap"
	"github.com/boms/backend/internal/config"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	"github.com/boms/backend/internal/infrastructure/logger"
	"github.com/boms/backend/internal/port"
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

	// The order jobs write events: deliver them at once, as the API does.
	pgPool.OnCommit(dispatcher.AfterCommit)

	pickupCodes := domainorder.NewPickupCodes(cfg.Order.PickupCodeSecret)
	unsubscribeTokens := domainpromotion.NewUnsubscribeTokens(cfg.Promotion.UnsubscribeSecret)
	stopEmails, err := startEmails(cfg.Mail, cfg.App.SiteURL, pickupCodes, unsubscribeTokens, redisClient.RDB(), pgPool, zlog)
	if err != nil {
		zlog.Fatal("email_init", zap.Error(err))
	}

	jobs := newJobs(cfg, pgPool, pickupCodes, zlog)

	zlog.Info("worker_started",
		zap.Duration("sweep_interval", cfg.Outbox.SweepInterval),
		zap.Duration("prune_interval", cfg.Outbox.PruneInterval),
		zap.Duration("order_job_interval", cfg.Order.JobInterval),
		zap.Int("email_concurrency", cfg.Mail.Concurrency),
	)
	run(ctx, dispatcher, postgresrepo.NewAssistantRepository(pgPool), jobs, redisrepo.NewWorkerHeartbeat(redisClient), cfg, zlog)
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
func startEmails(
	cfg config.MailConfig,
	siteURL string,
	codes domainorder.PickupCodes,
	unsubscribeTokens domainpromotion.UnsubscribeTokens,
	rdb *redis.Client,
	pool *postgresrepo.Pool,
	log *zap.Logger,
) (func(), error) {
	orderComposer, err := email.NewOrderComposer(siteURL)
	if err != nil {
		return nil, err
	}
	accountComposer, err := email.NewAccountComposer(siteURL)
	if err != nil {
		return nil, err
	}
	promotionComposer, err := email.NewPromotionComposer(siteURL)
	if err != nil {
		return nil, err
	}
	replyComposer, err := email.NewReplyComposer(siteURL)
	if err != nil {
		return nil, err
	}
	mailer := email.NewSMTPMailer(cfg)
	users := postgresrepo.NewUserRepository(pool)
	orderEmails := usecase.NewOrderEmailUsecase(postgresrepo.NewOrderRepository(pool), postgresrepo.NewPaymentRepository(pool),
		orderComposer, mailer, codes, log)
	accountEmails := usecase.NewAccountEmailUsecase(users, postgresrepo.NewUserTokenRepository(pool), accountComposer, mailer, log)
	replyEmails := usecase.NewReplyEmailUsecase(postgresrepo.NewConversationRepository(pool), replyComposer, mailer, log)
	promotions := usecase.NewPromotionEmailUsecase(postgresrepo.NewPromotionRepository(pool), queue.NewEmailQueue(rdb), pool,
		postgresrepo.NewOutboxRepository(pool), promotionComposer, mailer, unsubscribeTokens, log)
	// Order and account email and promotions each have their own server: a
	// promotion going out never takes a place an order email needs.
	emailServer := newEmailServer(rdb, cfg, cfg.Concurrency, queue.QueueEmail, log)
	promotionServer := newEmailServer(rdb, cfg, cfg.PromotionConcurrency, queue.QueuePromotion, log)
	emails := asynq.NewServeMux()
	emails.Handle(queue.TypeOrderEmail, queue.OrderEmailHandler(orderEmails))
	emails.Handle(queue.TypeAccountEmail, queue.AccountEmailHandler(accountEmails))
	emails.Handle(queue.TypeReplyEmail, queue.ReplyEmailHandler(replyEmails))
	promotionTasks := asynq.NewServeMux()
	promotionTasks.Handle(queue.TypePromotion, queue.PromotionHandler(promotions))
	promotionTasks.Handle(queue.TypePromotionEmail, queue.PromotionEmailHandler(promotions))
	if err := emailServer.Start(emails); err != nil {
		return nil, fmt.Errorf("start email queue: %w", err)
	}
	if err := promotionServer.Start(promotionTasks); err != nil {
		emailServer.Shutdown()
		return nil, fmt.Errorf("start promotion queue: %w", err)
	}
	return func() {
		// Neither takes a new task while the other finishes its own.
		emailServer.Stop()
		promotionServer.Stop()
		emailServer.Shutdown()
		promotionServer.Shutdown()
	}, nil
}

// newEmailServer sends the tasks of one queue, concurrency at a time.
func newEmailServer(rdb *redis.Client, cfg config.MailConfig, concurrency int, queueName string, log *zap.Logger) *asynq.Server {
	return asynq.NewServerFromRedisClient(rdb, asynq.Config{
		Concurrency: concurrency,
		Queues:      map[string]int{queueName: 1},
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
}

// outboxJobs is the outbox work the worker schedules.
type outboxJobs interface {
	Sweep(ctx context.Context, grace time.Duration, batch int32) (int, error)
	Prune(ctx context.Context, retention time.Duration) (int64, error)
}

// assistantSessions is the assistant's conversations, deleted once they pass
// their retention.
type assistantSessions interface {
	DeleteIdle(ctx context.Context, retention time.Duration) (int64, error)
}

// job is work the worker runs every ORDER_JOB_INTERVAL; run returns how many
// it handled — orders, conversations emailed about, or accounts cleared of
// their photos.
type job struct {
	name string
	run  func(ctx context.Context) (int, error)
}

// newJobs is the order work, the emails about replies left unread and the
// deletion of erased customers' reference photos.
func newJobs(cfg *config.Config, pool *postgresrepo.Pool, codes domainorder.PickupCodes, log *zap.Logger) []job {
	orders := postgresrepo.NewOrderRepository(pool)
	carts := postgresrepo.NewCartRepository(pool)
	discounts := postgresrepo.NewDiscountCodeRepository(pool)
	tickets := postgresrepo.NewTicketRepository(pool)
	paymentRepo := postgresrepo.NewPaymentRepository(pool)
	outbox := postgresrepo.NewOutboxRepository(pool)
	users := postgresrepo.NewUserRepository(pool)
	conversationRepo := postgresrepo.NewConversationRepository(pool)
	payments := usecase.NewPaymentUsecase(pool, orders, discounts, tickets, paymentRepo, paypal.New(cfg.PayPal), outbox,
		cfg.App.SiteURL, log)
	cartUC := usecase.NewCartUsecase(carts, postgresrepo.NewProductRepository(pool), postgresrepo.NewComboRepository(pool), discounts, cfg.Cloudinary)
	orderUC := usecase.NewOrderUsecase(users, orders, carts, discounts, cartUC, pool, outbox,
		postgresrepo.NewStoreSettingsRepository(pool), tickets, paymentRepo, payments, codes, conversationRepo,
		// The worker takes no one's acceptance of the policies.
		domainpolicy.Terms{})
	conversations := usecase.NewConversationUsecase(users, orders, conversationRepo, pool, outbox)
	jobs := []job{
		{name: "expire_unpaid", run: payments.ExpireOverdue},
		{name: "cancel_unreviewed", run: orderUC.CancelUnreviewed},
		{name: "make_refunds", run: payments.RefundDue},
		{name: "mark_no_shows", run: orderUC.MarkNoShows},
		{name: "flag_left_open", run: orderUC.FlagLeftOpen},
		{name: "email_unread_replies", run: conversations.EmailUnreadReplies},
	}
	// Cloudinary is required once deployed; without it in development no photo
	// was ever uploaded, and the erasures wait until it is set.
	if !cfg.Cloudinary.Enabled() {
		log.Warn("reference_photo_erasure_off", zap.String("reason", "cloudinary not configured"))
		return jobs
	}
	photos := usecase.NewReferencePhotoUsecase(postgresrepo.NewReferencePhotoRepository(pool), cloudinary.NewAdmin(cfg.Cloudinary), cfg.Cloudinary)
	return append(jobs, job{name: "erase_reference_photos", run: photos.EraseDue})
}

// run runs each job in a loop of its own, on its own interval, until ctx is
// cancelled: a job that is slow or stuck never holds the others back. Each run
// has its interval as its deadline, so the next one starts on time. Each round
// of an order job is recorded as the worker's last run.
func run(
	ctx context.Context,
	jobs outboxJobs,
	assistant assistantSessions,
	periodic []job,
	heartbeat port.WorkerHeartbeat,
	cfg *config.Config,
	log *zap.Logger,
) {
	var loops sync.WaitGroup
	every := func(interval time.Duration, now bool, work func(ctx context.Context)) {
		loops.Go(func() {
			// A restarted worker recovers at once instead of one interval later.
			if now {
				within(ctx, interval, work)
			}
			ticker := time.NewTicker(interval)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					within(ctx, interval, work)
				}
			}
		})
	}
	every(cfg.Outbox.SweepInterval, true, func(ctx context.Context) { sweep(ctx, jobs, cfg.Outbox, log) })
	every(cfg.Outbox.PruneInterval, false, func(ctx context.Context) { prune(ctx, jobs, assistant, cfg, log) })
	for _, j := range periodic {
		every(cfg.Order.JobInterval, false, func(runCtx context.Context) {
			// What a job could not do is still due, so its next run tries it again.
			moved, err := j.run(runCtx)
			if err != nil {
				log.Error("job_failed", zap.String("job", j.name), zap.Error(err))
			}
			if moved > 0 {
				log.Info("job_done", zap.String("job", j.name), zap.Int("handled", moved))
			}
			// The loop's context: a run cut at its deadline still shows the worker alive.
			if err := heartbeat.Beat(ctx, time.Now()); err != nil {
				log.Warn("worker_heartbeat_failed", zap.Error(err))
			}
		})
	}
	loops.Wait()
}

// within runs work with a deadline of its own.
func within(ctx context.Context, deadline time.Duration, work func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	work(ctx)
}

// prune deletes what has passed its retention: delivered outbox events, and
// conversations with the assistant idle longer than the privacy policy keeps them.
func prune(ctx context.Context, jobs outboxJobs, assistant assistantSessions, cfg *config.Config, log *zap.Logger) {
	if deleted, err := jobs.Prune(ctx, cfg.Outbox.Retention); err != nil {
		log.Error("outbox_prune_failed", zap.Error(err))
	} else if deleted > 0 {
		log.Info("outbox_pruned", zap.Int64("events", deleted))
	}
	if deleted, err := assistant.DeleteIdle(ctx, cfg.Assistant.Retention); err != nil {
		log.Error("assistant_prune_failed", zap.Error(err))
	} else if deleted > 0 {
		log.Info("assistant_pruned", zap.Int64("sessions", deleted))
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
