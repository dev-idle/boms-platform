package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boms/backend/internal/adapter/queue"
	postgresrepo "github.com/boms/backend/internal/adapter/repository/postgres"
	redisrepo "github.com/boms/backend/internal/adapter/repository/redis"
	"github.com/boms/backend/internal/config"
	v1 "github.com/boms/backend/internal/handler/v1"
	"github.com/boms/backend/internal/infrastructure/logger"
	"github.com/boms/backend/internal/middleware"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"

	_ "github.com/golang-jwt/jwt/v5" // pin JWT module until auth adapter exists

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
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

	rootCtx := context.Background()

	pgPool, err := postgresrepo.NewPool(rootCtx, cfg.Postgres)
	if err != nil {
		zlog.Fatal("postgres_init", zap.Error(err))
	}
	defer pgPool.Close()

	redisClient, err := redisrepo.NewClient(rootCtx, cfg.Redis)
	if err != nil {
		zlog.Fatal("redis_init", zap.Error(err))
	}
	defer func() { _ = redisClient.Close() }()

	var asynqClose func() error
	if cfg.Asynq.Enabled {
		client, err := queue.NewAsynqClient(cfg.Redis)
		if err != nil {
			zlog.Fatal("asynq_init", zap.Error(err))
		}
		asynqClose = client.Close
	}

	resources := []port.HealthResource{pgPool, redisClient}
	readinessTimeout := cfg.Postgres.HealthCheckTimeout
	if cfg.Redis.HealthCheckTimeout > readinessTimeout {
		readinessTimeout = cfg.Redis.HealthCheckTimeout
	}
	readiness := usecase.NewReadiness(resources, readinessTimeout+time.Second)
	health := v1.NewHealthHandler(readiness)

	app := newFiberApp(cfg, zlog)
	app.Get("/health", health.Live)
	app.Get("/ready", health.Ready)

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	go func() {
		zlog.Info("http_listen", zap.String("addr", addr), zap.String("env", cfg.App.Env))
		if err := app.Listen(addr); err != nil {
			zlog.Fatal("http_listen", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		zlog.Error("http_shutdown", zap.Error(err))
	}
	if asynqClose != nil {
		if err := asynqClose(); err != nil {
			zlog.Error("asynq_close", zap.Error(err))
		}
	}
	zlog.Info("shutdown_complete")
}

func newFiberApp(cfg *config.Config, log *zap.Logger) *fiber.App {
	fcfg := fiber.Config{
		AppName:               cfg.App.Name,
		ServerHeader:          "",
		StrictRouting:         true,
		ReadTimeout:           cfg.HTTP.ReadTimeout,
		WriteTimeout:          cfg.HTTP.WriteTimeout,
		IdleTimeout:           cfg.HTTP.IdleTimeout,
		BodyLimit:             cfg.HTTP.BodyLimit,
		DisableStartupMessage: cfg.App.Env == "production" || cfg.App.Env == "staging",
		EnablePrintRoutes:     cfg.App.Debug,
		ErrorHandler:          middleware.ErrorHandler(log),
	}

	if len(cfg.HTTP.TrustedProxies) > 0 {
		fcfg.EnableTrustedProxyCheck = true
		fcfg.TrustedProxies = cfg.HTTP.TrustedProxies
		fcfg.ProxyHeader = fiber.HeaderXForwardedFor
	}

	app := fiber.New(fcfg)

	app.Use(requestid.New())
	app.Use(middleware.SecurityHeaders(cfg.HTTP))
	app.Use(middleware.Recover(log))
	app.Use(middleware.RequestLogger(log))
	app.Use(middleware.CORS(cfg.CORS))
	app.Use(middleware.RateLimit(cfg.Rate))

	return app
}
