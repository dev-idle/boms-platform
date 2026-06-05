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
	meUC := usecase.NewMeUsecase(userRepo, customerProfileRepo, staffProfileRepo, adminProfileRepo, sessionStore, hasher, auditLogger, zlog)
	adminUserUC := usecase.NewAdminUserUsecase(userRepo, customerProfileRepo, staffProfileRepo, adminProfileRepo, sessionStore, pgPool, hasher, auditLogger, zlog)
	categoryRepo := postgresrepo.NewCategoryRepository(pgPool)
	productRepo := postgresrepo.NewProductRepository(pgPool)
	comboRepo := postgresrepo.NewComboRepository(pgPool)
	discountCodeRepo := postgresrepo.NewDiscountCodeRepository(pgPool)
	managerCategoryUC := usecase.NewManagerCategoryUsecase(categoryRepo, auditLogger, zlog)
	managerProductUC := usecase.NewManagerProductUsecase(productRepo, categoryRepo, auditLogger, zlog)
	managerComboUC := usecase.NewManagerComboUsecase(comboRepo, pgPool, auditLogger, zlog)
	managerDiscountCodeUC := usecase.NewManagerDiscountCodeUsecase(discountCodeRepo, auditLogger, zlog)
	catalogUC := usecase.NewCatalogUsecase(categoryRepo, productRepo, comboRepo)
	cartRepo := postgresrepo.NewCartRepository(pgPool)
	orderRepo := postgresrepo.NewOrderRepository(pgPool)
	cartUC := usecase.NewCartUsecase(cartRepo, productRepo, comboRepo, discountCodeRepo)
	orderUC := usecase.NewOrderUsecase(orderRepo, cartRepo, discountCodeRepo, cartUC, pgPool)

	if err := bootstrap.EnsureDevAdmin(rootCtx, cfg, userRepo, adminProfileRepo, hasher, pgPool); err != nil {
		zlog.Fatal("seed_admin", zap.Error(err))
	}

	authHandler := v1.NewAuthHandler(authUC, cfg)
	meHandler := v1.NewMeHandler(meUC)
	adminUserHandler := v1.NewAdminUserHandler(adminUserUC)
	managerCategoryHandler := v1.NewManagerCategoryHandler(managerCategoryUC)
	managerProductHandler := v1.NewManagerProductHandler(managerProductUC)
	managerComboHandler := v1.NewManagerComboHandler(managerComboUC)
	managerDiscountCodeHandler := v1.NewManagerDiscountCodeHandler(managerDiscountCodeUC)
	catalogHandler := v1.NewCatalogHandler(catalogUC)
	cartHandler := v1.NewCartHandler(cartUC)
	orderHandler := v1.NewOrderHandler(orderUC)

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
	// Health probes are intentionally unauthenticated and bypass the proxy secret.
	app.Get("/health", health.Live)
	app.Get("/ready", health.Ready)

	// All /api/v1/* traffic must originate from the Next.js proxy (verified by shared secret).
	apiV1 := app.Group("/api/v1", middleware.RequireInternalSecret(cfg.HTTP.InternalSecret))

	rdb := redisClient.RDB()
	authGroup := apiV1.Group("/auth")
	authGroup.Post("/register", middleware.AuthAttemptRateLimit(rdb, cfg.RateRedis), authHandler.Register)
	authGroup.Post("/login", middleware.AuthAttemptRateLimit(rdb, cfg.RateRedis), authHandler.Login)
	authGroup.Post("/refresh", middleware.AuthRefreshRateLimit(rdb, cfg.RateRedis), authHandler.Refresh)
	authGroup.Post("/logout", middleware.AuthLogoutRateLimit(rdb, cfg.RateRedis), middleware.OptionalAuth(tokenSigner), authHandler.Logout)

	passwordChanged := middleware.RequirePasswordChanged(sessionStore)

	apiV1.Get("/me", middleware.RequireAuth(tokenSigner), meHandler.Get)
	apiV1.Patch("/me", middleware.RequireAuthWithSession(tokenSigner, sessionStore), passwordChanged, meHandler.Patch)
	apiV1.Patch("/me/password", middleware.RequireAuthWithSession(tokenSigner, sessionStore), meHandler.PatchPassword)
	apiV1.Delete("/me", middleware.RequireAuthWithSession(tokenSigner, sessionStore), passwordChanged, meHandler.Delete)

	adminRead := apiV1.Group(
		"/admin/users",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleAdmin),
		passwordChanged,
	)
	adminRead.Get("", adminUserHandler.List)
	adminRead.Get("/:id", adminUserHandler.Get)

	adminWrite := apiV1.Group(
		"/admin/users",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleAdmin),
		passwordChanged,
		middleware.AdminWriteRateLimit(rdb, cfg.RateRedis),
	)
	adminWrite.Post("", adminUserHandler.Create)
	adminWrite.Patch("/:id", adminUserHandler.PatchProfile)
	adminWrite.Patch("/:id/role", adminUserHandler.PatchRole)
	adminWrite.Patch("/:id/disable", adminUserHandler.PatchDisable)
	adminWrite.Post("/:id/revoke-sessions", adminUserHandler.RevokeSessions)

	catalogRead := apiV1.Group(
		"/catalog",
		middleware.RequireAuth(tokenSigner),
		middleware.RequireRole(domainuser.RoleCustomer),
	)
	catalogRead.Get("/categories", catalogHandler.ListCategories)
	catalogRead.Get("/products", catalogHandler.ListProducts)
	catalogRead.Get("/products/:id", catalogHandler.GetProduct)
	catalogRead.Get("/combos", catalogHandler.ListCombos)
	catalogRead.Get("/combos/:id", catalogHandler.GetCombo)

	customerSession := apiV1.Group(
		"",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleCustomer),
		passwordChanged,
	)
	customerSession.Get("/cart", cartHandler.Get)
	customerSession.Post("/cart/items", cartHandler.AddItem)
	customerSession.Patch("/cart/items/:id", cartHandler.UpdateItem)
	customerSession.Delete("/cart/items/:id", cartHandler.RemoveItem)
	customerSession.Put("/cart/discount", cartHandler.ApplyDiscount)
	customerSession.Delete("/cart/discount", cartHandler.RemoveDiscount)
	customerSession.Post("/orders/checkout", orderHandler.Checkout)
	customerSession.Get("/orders", orderHandler.List)
	customerSession.Get("/orders/:id", orderHandler.Get)

	managerRead := apiV1.Group(
		"/manager",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleManager),
		passwordChanged,
	)
	managerRead.Get("/categories", managerCategoryHandler.List)
	managerRead.Get("/categories/:id", managerCategoryHandler.Get)
	managerRead.Get("/products", managerProductHandler.List)
	managerRead.Get("/products/:id", managerProductHandler.Get)
	managerRead.Get("/combos", managerComboHandler.List)
	managerRead.Get("/combos/:id", managerComboHandler.Get)
	managerRead.Get("/discount-codes", managerDiscountCodeHandler.List)
	managerRead.Get("/discount-codes/:id", managerDiscountCodeHandler.Get)

	managerWrite := apiV1.Group(
		"/manager",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleManager),
		passwordChanged,
		middleware.ManagerWriteRateLimit(rdb, cfg.RateRedis),
	)
	managerWrite.Post("/categories", managerCategoryHandler.Create)
	managerWrite.Patch("/categories/:id", managerCategoryHandler.Patch)
	managerWrite.Delete("/categories/:id", managerCategoryHandler.Delete)
	managerWrite.Post("/products", managerProductHandler.Create)
	managerWrite.Patch("/products/:id", managerProductHandler.Patch)
	managerWrite.Delete("/products/:id", managerProductHandler.Delete)
	managerWrite.Post("/combos", managerComboHandler.Create)
	managerWrite.Patch("/combos/:id", managerComboHandler.Patch)
	managerWrite.Delete("/combos/:id", managerComboHandler.Delete)
	managerWrite.Post("/discount-codes", managerDiscountCodeHandler.Create)
	managerWrite.Patch("/discount-codes/:id", managerDiscountCodeHandler.Patch)
	managerWrite.Delete("/discount-codes/:id", managerDiscountCodeHandler.Delete)

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

	// Order matters: requestid → recover (catches panics from everything below) →
	// request meta + headers → logger → cors → rate limit.
	app.Use(requestid.New())
	app.Use(middleware.Recover(log))
	app.Use(middleware.AttachRequestMeta())
	app.Use(middleware.SecurityHeaders(cfg.HTTP))
	app.Use(middleware.RequestLogger(log))
	app.Use(middleware.CORS(cfg.CORS))
	app.Use(middleware.RateLimit(cfg.Rate))

	return app
}
