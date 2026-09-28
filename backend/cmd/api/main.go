package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/boms/backend/internal/adapter/eventbus"
	"github.com/boms/backend/internal/adapter/queue"
	"github.com/boms/backend/internal/adapter/realtime"
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
	"github.com/boms/backend/internal/service/eventdispatch"
	"github.com/boms/backend/internal/usecase"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/net/netutil"
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
	meUC := usecase.NewMeUsecase(userRepo, customerProfileRepo, staffProfileRepo, adminProfileRepo, sessionStore, pgPool, hasher, auditLogger, zlog)
	adminUserUC := usecase.NewAdminUserUsecase(userRepo, customerProfileRepo, staffProfileRepo, adminProfileRepo, sessionStore, pgPool, hasher, auditLogger, auditLogRepo, zlog)
	categoryRepo := postgresrepo.NewCategoryRepository(pgPool)
	productRepo := postgresrepo.NewProductRepository(pgPool)
	comboRepo := postgresrepo.NewComboRepository(pgPool)
	discountCodeRepo := postgresrepo.NewDiscountCodeRepository(pgPool)
	managerCategoryUC := usecase.NewManagerCategoryUsecase(categoryRepo, auditLogger, zlog)
	managerMediaUC := usecase.NewManagerMediaUsecase(cfg.Cloudinary)
	managerProductUC := usecase.NewManagerProductUsecase(productRepo, categoryRepo, pgPool, auditLogger, cfg.Cloudinary, zlog)
	managerComboUC := usecase.NewManagerComboUsecase(comboRepo, pgPool, auditLogger, cfg.Cloudinary, zlog)
	managerDiscountCodeUC := usecase.NewManagerDiscountCodeUsecase(discountCodeRepo, auditLogger, zlog)
	catalogUC := usecase.NewCatalogUsecase(categoryRepo, productRepo, comboRepo)
	cartRepo := postgresrepo.NewCartRepository(pgPool)
	orderRepo := postgresrepo.NewOrderRepository(pgPool)
	cartUC := usecase.NewCartUsecase(cartRepo, productRepo, comboRepo, discountCodeRepo)
	outboxRepo := postgresrepo.NewOutboxRepository(pgPool)
	eventDispatcher := eventdispatch.New(outboxRepo, eventbus.NewRedisPublisher(redisClient.RDB()), pgPool, zlog, cfg.Outbox.DispatchTimeout)
	pgPool.OnCommit(eventDispatcher.AfterCommit)
	orderUC := usecase.NewOrderUsecase(orderRepo, cartRepo, discountCodeRepo, cartUC, pgPool, outboxRepo)
	staffOrderUC := usecase.NewStaffOrderUsecase(orderRepo, pgPool, outboxRepo, auditLogger, zlog)
	bakerOrderUC := usecase.NewBakerOrderUsecase(orderRepo, pgPool, outboxRepo, auditLogger, zlog)
	realtimeTickets := redisrepo.NewRealtimeTicketStore(redisClient)
	realtimeUC := usecase.NewRealtimeUsecase(realtimeTickets, cfg.Realtime.PublicURL, cfg.Realtime.TicketTTL)

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
	managerMediaHandler := v1.NewManagerMediaHandler(managerMediaUC)
	catalogHandler := v1.NewCatalogHandler(catalogUC)
	cartHandler := v1.NewCartHandler(cartUC)
	orderHandler := v1.NewOrderHandler(orderUC)
	staffOrderHandler := v1.NewStaffOrderHandler(staffOrderUC)
	bakerOrderHandler := v1.NewBakerOrderHandler(bakerOrderUC)
	realtimeHandler := v1.NewRealtimeHandler(realtimeUC)

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
	selfWrite := middleware.SelfWriteRateLimit(rdb, cfg.RateRedis)

	apiV1.Get("/me", middleware.RequireAuth(tokenSigner), meHandler.Get)
	apiV1.Patch("/me", middleware.RequireAuthWithSession(tokenSigner, sessionStore), passwordChanged, selfWrite, meHandler.Patch)
	apiV1.Patch("/me/password", middleware.RequireAuthWithSession(tokenSigner, sessionStore), selfWrite, meHandler.PatchPassword)
	apiV1.Delete("/me", middleware.RequireAuthWithSession(tokenSigner, sessionStore), passwordChanged, selfWrite, meHandler.Delete)

	// Any signed-in session may open the push socket: the ticket records its role,
	// and the realtime listener picks the channels from it.
	apiV1.Post(
		"/realtime/tickets",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		passwordChanged,
		middleware.RealtimeTicketRateLimit(rdb, cfg.RateRedis),
		realtimeHandler.IssueTicket,
	)

	adminRead := apiV1.Group(
		"/admin/users",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleAdmin),
		passwordChanged,
	)
	adminRead.Get("", adminUserHandler.List)
	adminRead.Get("/:id/activity", adminUserHandler.ListActivity)
	adminRead.Get("/:id", adminUserHandler.Get)

	adminEmployeeCodes := apiV1.Group(
		"/admin/employee-codes",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleAdmin),
		passwordChanged,
	)
	adminEmployeeCodes.Get("/next", adminUserHandler.GetNextEmployeeCode)

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
	adminWrite.Patch("/:id/enable", adminUserHandler.PatchEnable)
	adminWrite.Post("/:id/reset-password", adminUserHandler.PostResetPassword)
	adminWrite.Post("/:id/revoke-sessions", adminUserHandler.RevokeSessions)

	// Public storefront read — proxy secret only; no JWT (guest browse).
	catalogRead := apiV1.Group("/catalog")
	catalogRead.Get("/categories", catalogHandler.ListCategories)
	catalogRead.Get("/products", catalogHandler.ListProducts)
	catalogRead.Get("/products/:id", catalogHandler.GetProduct)
	catalogRead.Get("/combos", catalogHandler.ListCombos)
	catalogRead.Get("/combos/:id", catalogHandler.GetCombo)

	customerCart := customerSessionGroup(apiV1, "/cart", tokenSigner, sessionStore, passwordChanged)
	customerCart.Get("", cartHandler.Get)
	customerCart.Post("/items", cartHandler.AddItem)
	customerCart.Patch("/items/:id", cartHandler.UpdateItem)
	customerCart.Delete("/items/:id", cartHandler.RemoveItem)
	customerCart.Put("/discount", middleware.DiscountAttemptRateLimit(rdb, cfg.RateRedis), cartHandler.ApplyDiscount)
	customerCart.Delete("/discount", cartHandler.RemoveDiscount)

	customerOrders := customerSessionGroup(apiV1, "/orders", tokenSigner, sessionStore, passwordChanged)
	customerOrders.Post("/checkout", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), orderHandler.Checkout)
	customerOrders.Get("", orderHandler.List)
	customerOrders.Get("/:id", orderHandler.Get)

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
	managerRead.Get(
		"/media/cloudinary-signature",
		middleware.ManagerMediaRateLimit(rdb, cfg.RateRedis),
		managerMediaHandler.GetCloudinaryUploadSignature,
	)

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

	staffOrders := apiV1.Group(
		"/staff",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleStaff),
		passwordChanged,
	)
	staffOrders.Get("/orders", staffOrderHandler.List)
	staffOrders.Get("/orders/:id", staffOrderHandler.Get)
	staffOrders.Patch("/orders/:id/status", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), staffOrderHandler.PatchStatus)

	bakerOrders := apiV1.Group(
		"/baker",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleBaker),
		passwordChanged,
	)
	bakerOrders.Get("/production", bakerOrderHandler.List)
	bakerOrders.Get("/production/:id", bakerOrderHandler.Get)
	bakerOrders.Patch("/production/:id/status", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), bakerOrderHandler.PatchStatus)

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	listenCfg := fiber.ListenConfig{
		DisableStartupMessage: cfg.App.Env == "production" || cfg.App.Env == "staging",
		EnablePrintRoutes:     cfg.App.Debug,
	}
	go func() {
		zlog.Info("http_listen", zap.String("addr", addr), zap.String("env", cfg.App.Env))
		if err := app.Listen(addr, listenCfg); err != nil {
			zlog.Fatal("http_listen", zap.Error(err))
		}
	}()
	stopRealtime := startRealtime(rootCtx, cfg, rdb, realtimeTickets, sessionStore, zlog)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		zlog.Error("http_shutdown", zap.Error(err))
	}
	stopRealtime(shutdownCtx)
	// Deliveries started by requests that just finished still need the pool.
	eventDispatcher.Wait(shutdownCtx)
	if asynqClose != nil {
		if err := asynqClose(); err != nil {
			zlog.Error("asynq_close", zap.Error(err))
		}
	}
	zlog.Info("shutdown_complete")
}

func newFiberApp(cfg *config.Config, log *zap.Logger) *fiber.App {
	fcfg := fiber.Config{
		AppName:       cfg.App.Name,
		ServerHeader:  "",
		StrictRouting: true,
		ReadTimeout:   cfg.HTTP.ReadTimeout,
		WriteTimeout:  cfg.HTTP.WriteTimeout,
		IdleTimeout:   cfg.HTTP.IdleTimeout,
		BodyLimit:     cfg.HTTP.BodyLimit,
		ErrorHandler:  middleware.ErrorHandler(log),
	}

	trustProxies(&fcfg, cfg.HTTP.TrustedProxies)

	app := fiber.New(fcfg)

	// Order matters: requestid → recover (catches panics from everything below) →
	// request deadline → request meta + headers → logger → cors → rate limit.
	app.Use(requestid.New(requestid.Config{Generator: uuid.NewString}))
	app.Use(middleware.Recover(log))
	app.Use(middleware.RequestTimeout(cfg.HTTP.RequestTimeout))
	app.Use(middleware.AttachRequestMeta())
	app.Use(middleware.SecurityHeaders(cfg.HTTP))
	app.Use(middleware.RequestLogger(log))
	app.Use(middleware.CORS(cfg.CORS))
	app.Use(middleware.RateLimit(cfg.Rate))

	return app
}

// trustProxies lets c.IP() read X-Forwarded-For only when the request comes from
// one of the listed proxies. IP validation walks the header right to left past
// those proxies and falls back to the socket address, so a caller cannot choose
// the address its rate-limit bucket and audit rows are keyed by.
func trustProxies(fcfg *fiber.Config, proxies []string) {
	if len(proxies) == 0 {
		return
	}
	fcfg.TrustProxy = true
	fcfg.TrustProxyConfig = fiber.TrustProxyConfig{Proxies: proxies}
	fcfg.ProxyHeader = fiber.HeaderXForwardedFor
	fcfg.EnableIPValidation = true
}

// customerSessionGroup scopes customer session auth to an explicit sub-path.
// Never call apiV1.Group("") — empty prefix registers middleware on /api/v1 and
// leaks RequireRole(customer) onto every route declared later in main.
func customerSessionGroup(
	parent fiber.Router,
	path string,
	tokenSigner port.TokenSigner,
	sessions port.SessionStore,
	passwordChanged fiber.Handler,
) fiber.Router {
	return parent.Group(
		path,
		middleware.RequireAuthWithSession(tokenSigner, sessions),
		middleware.RequireRole(domainuser.RoleCustomer),
		passwordChanged,
	)
}

// realtimeMaxHeaderBytes bounds a WebSocket handshake. Browsers send the site's
// cookies along, so it leaves room for them, but nothing near the 1 MB default.
const realtimeMaxHeaderBytes = 32 << 10

// startRealtime opens the push-only WebSocket listener and feeds it every event
// on the bus. It runs apart from the API because browsers connect to it
// directly, not through the BFF. The returned stop closes the listener, then
// the open sockets, then the bus subscription.
func startRealtime(
	ctx context.Context,
	cfg *config.Config,
	rdb *goredis.Client,
	tickets port.RealtimeTicketStore,
	sessions port.SessionStore,
	log *zap.Logger,
) func(context.Context) {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", cfg.Realtime.Addr)
	if err != nil {
		log.Fatal("realtime_listen", zap.Error(err))
	}
	hub := realtime.NewHub(cfg.Realtime.MaxConnsPerUser)
	server := realtime.NewServer(hub, tickets, sessions, cfg.Realtime, log)
	// The timeouts bound a handshake from its first byte to its answer. An
	// upgraded socket is hijacked, which clears them; from then on its own
	// write timeout and pings keep it honest.
	handshake := cfg.Realtime.HandshakeTimeout
	httpServer := &http.Server{
		Handler:           server.Handler(),
		ReadHeaderTimeout: handshake,
		ReadTimeout:       handshake,
		WriteTimeout:      handshake,
		IdleTimeout:       handshake,
		MaxHeaderBytes:    realtimeMaxHeaderBytes,
		ErrorLog:          zap.NewStdLog(log),
	}

	busCtx, stopBus := context.WithCancel(context.Background())
	busDone := make(chan struct{})
	go func() {
		defer close(busDone)
		eventbus.Subscribe(busCtx, rdb, log, hub.Deliver)
	}()
	go func() {
		log.Info("realtime_listen", zap.String("addr", cfg.Realtime.Addr))
		// A hijacked socket keeps its slot until it closes, so this caps open
		// sockets as well as handshakes in flight.
		err := httpServer.Serve(netutil.LimitListener(listener, cfg.Realtime.MaxConns))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("realtime_listen", zap.Error(err))
		}
	}()

	return func(ctx context.Context) {
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Error("realtime_shutdown", zap.Error(err))
		}
		server.Close(ctx)
		stopBus()
		<-busDone
	}
}
