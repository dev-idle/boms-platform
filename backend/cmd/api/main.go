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
	"github.com/boms/backend/internal/bootstrap"
	"github.com/boms/backend/internal/config"
	domainuser "github.com/boms/backend/internal/domain/user"
	v1 "github.com/boms/backend/internal/handler/v1"
	"github.com/boms/backend/internal/infrastructure/crypto"
	jwtinfra "github.com/boms/backend/internal/infrastructure/jwt"
	"github.com/boms/backend/internal/infrastructure/logger"
	"github.com/boms/backend/internal/middleware"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	"github.com/boms/backend/internal/usecase"

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

	hasher := crypto.NewArgon2Hasher(cfg.Argon2)

	if err := jwtinfra.EnsureDevSeed(&cfg.JWT, cfg.App.Env == "development"); err != nil {
		zlog.Fatal("jwt_seed", zap.Error(err))
	}
	tokenSigner, err := jwtinfra.NewEdDSASigner(cfg.JWT)
	if err != nil {
		zlog.Fatal("jwt_signer", zap.Error(err))
	}

	userRepo := postgresrepo.NewUserRepository(pgPool)
	customerProfileRepo := postgresrepo.NewCustomerProfileRepository(pgPool)
	staffProfileRepo := postgresrepo.NewStaffProfileRepository(pgPool)
	adminProfileRepo := postgresrepo.NewAdminProfileRepository(pgPool)
	auditLogRepo := postgresrepo.NewAuditLogRepository(pgPool)
	sessionStore := redisrepo.NewSessionStore(redisClient, cfg.Session.TTL)
	auditLogger := auditlogger.NewService(auditLogRepo)

	authUC, err := usecase.NewAuthUsecase(userRepo, customerProfileRepo, pgPool, sessionStore, hasher, tokenSigner, zlog)
	if err != nil {
		zlog.Fatal("auth_usecase", zap.Error(err))
	}
	meUC := usecase.NewMeUsecase(userRepo, customerProfileRepo, staffProfileRepo, adminProfileRepo, sessionStore, hasher, auditLogger)
	adminUserUC := usecase.NewAdminUserUsecase(userRepo, customerProfileRepo, staffProfileRepo, adminProfileRepo, sessionStore, pgPool, hasher, auditLogger)

	if err := bootstrap.EnsureDevAdmin(rootCtx, cfg, userRepo, adminProfileRepo, hasher, pgPool); err != nil {
		zlog.Fatal("seed_admin", zap.Error(err))
	}

	authHandler := v1.NewAuthHandler(authUC, cfg)
	meHandler := v1.NewMeHandler(meUC)
	adminUserHandler := v1.NewAdminUserHandler(adminUserUC)

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
	readiness := usecase.NewReadiness(resources, readinessTimeout+time.Second, zlog)
	health := v1.NewHealthHandler(readiness)

	app := newFiberApp(cfg, zlog)
	app.Get("/health", health.Live)
	app.Get("/ready", health.Ready)

	rdb := redisClient.RDB()
	authGroup := app.Group("/api/v1/auth")
	authGroup.Post("/register", middleware.AuthAttemptRateLimit(rdb), authHandler.Register)
	authGroup.Post("/login", middleware.AuthAttemptRateLimit(rdb), authHandler.Login)
	authGroup.Post("/refresh", middleware.AuthRefreshRateLimit(rdb), authHandler.Refresh)
	authGroup.Post("/logout", middleware.AuthLogoutRateLimit(rdb), middleware.OptionalAuth(tokenSigner), authHandler.Logout)

	passwordChanged := middleware.RequirePasswordChanged(userRepo)

	app.Get("/api/v1/me", middleware.RequireAuth(tokenSigner), meHandler.Get)
	app.Patch("/api/v1/me", middleware.RequireAuthWithSession(tokenSigner, sessionStore), passwordChanged, meHandler.Patch)
	app.Patch("/api/v1/me/password", middleware.RequireAuthWithSession(tokenSigner, sessionStore), meHandler.PatchPassword)
	app.Delete("/api/v1/me", middleware.RequireAuthWithSession(tokenSigner, sessionStore), passwordChanged, meHandler.Delete)

	adminRead := app.Group(
		"/api/v1/admin/users",
		middleware.RequireAuth(tokenSigner),
		middleware.RequireRole(domainuser.RoleAdmin),
		passwordChanged,
	)
	adminRead.Get("", adminUserHandler.List)
	adminRead.Get("/:id", adminUserHandler.Get)

	adminWrite := app.Group(
		"/api/v1/admin/users",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleAdmin),
		passwordChanged,
		middleware.AdminWriteRateLimit(rdb),
	)
	adminWrite.Post("", adminUserHandler.Create)
	adminWrite.Patch("/:id", adminUserHandler.PatchProfile)
	adminWrite.Patch("/:id/role", adminUserHandler.PatchRole)
	adminWrite.Patch("/:id/disable", adminUserHandler.PatchDisable)
	adminWrite.Post("/:id/revoke-sessions", adminUserHandler.RevokeSessions)

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
	app.Use(middleware.AttachRequestMeta())
	app.Use(middleware.SecurityHeaders(cfg.HTTP))
	app.Use(middleware.Recover(log))
	app.Use(middleware.RequestLogger(log))
	app.Use(middleware.CORS(cfg.CORS))
	app.Use(middleware.RateLimit(cfg.Rate))

	return app
}
