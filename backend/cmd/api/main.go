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
	"github.com/boms/backend/internal/adapter/paypal"
	"github.com/boms/backend/internal/adapter/realtime"
	postgresrepo "github.com/boms/backend/internal/adapter/repository/postgres"
	redisrepo "github.com/boms/backend/internal/adapter/repository/redis"
	"github.com/boms/backend/internal/bootstrap"
	"github.com/boms/backend/internal/config"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpromotion "github.com/boms/backend/internal/domain/promotion"
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
	outboxRepo := postgresrepo.NewOutboxRepository(pgPool)
	userTokenRepo := postgresrepo.NewUserTokenRepository(pgPool)

	authUC, err := usecase.NewAuthUsecase(userRepo, customerProfileRepo, pgPool, outboxRepo, sessionStore, hasher, tokenSigner, zlog)
	if err != nil {
		zlog.Fatal("auth_usecase", zap.Error(err))
	}
	meUC := usecase.NewMeUsecase(userRepo, customerProfileRepo, staffProfileRepo, adminProfileRepo, userTokenRepo, sessionStore, pgPool, hasher, auditLogger, zlog)
	adminUserUC := usecase.NewAdminUserUsecase(
		userRepo, customerProfileRepo, staffProfileRepo, adminProfileRepo, userTokenRepo, sessionStore, pgPool, hasher, auditLogger, auditLogRepo, zlog,
	)
	categoryRepo := postgresrepo.NewCategoryRepository(pgPool)
	productRepo := postgresrepo.NewProductRepository(pgPool)
	comboRepo := postgresrepo.NewComboRepository(pgPool)
	discountCodeRepo := postgresrepo.NewDiscountCodeRepository(pgPool)
	managerCategoryUC := usecase.NewManagerCategoryUsecase(categoryRepo, auditLogger, zlog)
	mediaUC := usecase.NewMediaUsecase(cfg.Cloudinary, userRepo)
	managerProductUC := usecase.NewManagerProductUsecase(productRepo, categoryRepo, pgPool, auditLogger, cfg.Cloudinary, zlog)
	managerComboUC := usecase.NewManagerComboUsecase(comboRepo, pgPool, auditLogger, cfg.Cloudinary, zlog)
	managerDiscountCodeUC := usecase.NewManagerDiscountCodeUsecase(discountCodeRepo, auditLogger, zlog)
	catalogUC := usecase.NewCatalogUsecase(categoryRepo, productRepo, comboRepo)
	cartRepo := postgresrepo.NewCartRepository(pgPool)
	orderRepo := postgresrepo.NewOrderRepository(pgPool)
	cartUC := usecase.NewCartUsecase(cartRepo, productRepo, comboRepo, discountCodeRepo, cfg.Cloudinary)
	storeSettingsRepo := postgresrepo.NewStoreSettingsRepository(pgPool)
	eventDispatcher := eventdispatch.New(outboxRepo, bootstrap.EventPublisher(redisClient.RDB()), pgPool, zlog, cfg.Outbox.DispatchTimeout)
	pgPool.OnCommit(eventDispatcher.AfterCommit)
	ticketRepo := postgresrepo.NewTicketRepository(pgPool)
	paymentRepo := postgresrepo.NewPaymentRepository(pgPool)
	paymentUC := usecase.NewPaymentUsecase(pgPool, orderRepo, discountCodeRepo, ticketRepo, paymentRepo, paypal.New(cfg.PayPal), outboxRepo, cfg.App.SiteURL, zlog)
	pickupCodes := domainorder.NewPickupCodes(cfg.Order.PickupCodeSecret)
	quota := redisrepo.NewQuota(redisClient)
	conversationRepo := postgresrepo.NewConversationRepository(pgPool)
	savedProductRepo := postgresrepo.NewSavedProductRepository(pgPool)
	reviewRepo := postgresrepo.NewReviewRepository(pgPool)
	promotionRepo := postgresrepo.NewPromotionRepository(pgPool)
	orderUC := usecase.NewOrderUsecase(userRepo, orderRepo, cartRepo, discountCodeRepo, cartUC, pgPool, outboxRepo, storeSettingsRepo, ticketRepo, paymentRepo, paymentUC, pickupCodes,
		conversationRepo)
	storeUC := usecase.NewStoreUsecase(storeSettingsRepo, orderRepo)
	accountErasureUC := usecase.NewAccountErasureUsecase(
		pgPool, userRepo, customerProfileRepo, cartRepo, orderRepo, conversationRepo, savedProductRepo, reviewRepo, auditLogRepo, userTokenRepo, sessionStore, auditLogger, hasher,
	)
	emailVerificationUC := usecase.NewEmailVerificationUsecase(pgPool, userRepo, userTokenRepo, outboxRepo, auditLogger)
	passwordResetUC := usecase.NewPasswordResetUsecase(
		pgPool, userRepo, userTokenRepo, outboxRepo, sessionStore, hasher, quota,
		port.QuotaLimit{Max: cfg.RateRedis.PasswordResetAccountMax, Window: cfg.RateRedis.PasswordResetAccountWindow},
		auditLogger, zlog,
	)
	dataExportUC := usecase.NewDataExportUsecase(
		userRepo, customerProfileRepo, staffProfileRepo, adminProfileRepo, orderRepo, conversationRepo, savedProductRepo, reviewRepo, cartRepo, sessionStore, auditLogRepo,
	)
	adminStoreSettingsUC := usecase.NewAdminStoreSettingsUsecase(storeSettingsRepo, pgPool, outboxRepo, auditLogger, zlog)
	staffOrderUC := usecase.NewStaffOrderUsecase(userRepo, orderRepo, ticketRepo, pgPool, outboxRepo, auditLogger, zlog, paymentRepo, discountCodeRepo, storeSettingsRepo, cartUC,
		pickupCodes, quota, port.QuotaLimit{Max: cfg.RateRedis.PickupCodeMax, Window: cfg.RateRedis.PickupCodeWindow})
	staffProductUC := usecase.NewStaffProductUsecase(productRepo, pgPool, outboxRepo, auditLogger, zlog)
	conversationUC := usecase.NewConversationUsecase(userRepo, orderRepo, conversationRepo, pgPool, outboxRepo)
	staffConversationUC := usecase.NewStaffConversationUsecase(userRepo, orderRepo, conversationRepo, pgPool, outboxRepo)
	savedProductUC := usecase.NewSavedProductUsecase(userRepo, savedProductRepo, pgPool)
	reviewUC := usecase.NewReviewUsecase(userRepo, orderRepo, reviewRepo, pgPool, outboxRepo)
	managerReviewUC := usecase.NewManagerReviewUsecase(reviewRepo, pgPool, outboxRepo, auditLogger, zlog)
	managerPromotionUC := usecase.NewManagerPromotionUsecase(promotionRepo, pgPool, outboxRepo, auditLogger, zlog)
	managerIncidentUC := usecase.NewManagerIncidentUsecase(orderRepo)
	managerEngagementUC := usecase.NewManagerEngagementUsecase(postgresrepo.NewEngagementRepository(pgPool))
	managerSalesReportUC := usecase.NewManagerSalesReportUsecase(postgresrepo.NewSalesReportRepository(pgPool))
	unsubscribeUC := usecase.NewUnsubscribeUsecase(userRepo, customerProfileRepo, pgPool, auditLogger,
		domainpromotion.NewUnsubscribeTokens(cfg.Promotion.UnsubscribeSecret))
	staffTicketUC := usecase.NewStaffTicketUsecase(orderRepo, ticketRepo, pgPool, outboxRepo, auditLogger, zlog)
	bakerTicketUC := usecase.NewBakerTicketUsecase(orderRepo, ticketRepo, pgPool, outboxRepo, auditLogger, zlog)
	realtimeTickets := redisrepo.NewRealtimeTicketStore(redisClient)
	realtimeUC := usecase.NewRealtimeUsecase(realtimeTickets, cfg.Realtime.PublicURL, cfg.Realtime.TicketTTL)

	if err := bootstrap.EnsureDevAdmin(rootCtx, cfg, userRepo, adminProfileRepo, hasher, pgPool); err != nil {
		zlog.Fatal("seed_admin", zap.Error(err))
	}

	authHandler := v1.NewAuthHandler(authUC, cfg)
	meHandler := v1.NewMeHandler(meUC)
	dataExportHandler := v1.NewDataExportHandler(dataExportUC)
	accountErasureHandler := v1.NewAccountErasureHandler(accountErasureUC)
	emailVerificationHandler := v1.NewEmailVerificationHandler(emailVerificationUC)
	passwordResetHandler := v1.NewPasswordResetHandler(passwordResetUC)
	adminUserHandler := v1.NewAdminUserHandler(adminUserUC)
	managerCategoryHandler := v1.NewManagerCategoryHandler(managerCategoryUC)
	managerProductHandler := v1.NewManagerProductHandler(managerProductUC)
	managerComboHandler := v1.NewManagerComboHandler(managerComboUC)
	managerDiscountCodeHandler := v1.NewManagerDiscountCodeHandler(managerDiscountCodeUC)
	mediaHandler := v1.NewMediaHandler(mediaUC)
	catalogHandler := v1.NewCatalogHandler(catalogUC)
	cartHandler := v1.NewCartHandler(cartUC)
	orderHandler := v1.NewOrderHandler(orderUC)
	paymentHandler := v1.NewPaymentHandler(paymentUC)
	staffOrderHandler := v1.NewStaffOrderHandler(staffOrderUC)
	staffTicketHandler := v1.NewStaffTicketHandler(staffTicketUC)
	staffProductHandler := v1.NewStaffProductHandler(staffProductUC)
	conversationHandler := v1.NewConversationHandler(conversationUC)
	staffConversationHandler := v1.NewStaffConversationHandler(staffConversationUC)
	savedProductHandler := v1.NewSavedProductHandler(savedProductUC)
	reviewHandler := v1.NewReviewHandler(reviewUC)
	managerReviewHandler := v1.NewManagerReviewHandler(managerReviewUC)
	managerPromotionHandler := v1.NewManagerPromotionHandler(managerPromotionUC)
	managerIncidentHandler := v1.NewManagerIncidentHandler(managerIncidentUC)
	managerEngagementHandler := v1.NewManagerEngagementHandler(managerEngagementUC)
	managerSalesReportHandler := v1.NewManagerSalesReportHandler(managerSalesReportUC)
	unsubscribeHandler := v1.NewUnsubscribeHandler(unsubscribeUC)
	bakerTicketHandler := v1.NewBakerTicketHandler(bakerTicketUC)
	realtimeHandler := v1.NewRealtimeHandler(realtimeUC)
	storeHandler := v1.NewStoreHandler(storeUC)
	adminStoreSettingsHandler := v1.NewAdminStoreSettingsHandler(adminStoreSettingsUC)

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
	// Emailed links work without a session: they are often opened on another device.
	authGroup.Post("/verify-email", middleware.AuthLinkRateLimit(rdb, cfg.RateRedis), emailVerificationHandler.Verify)
	authGroup.Post("/password-reset/request", middleware.PasswordResetRateLimit(rdb, cfg.RateRedis), passwordResetHandler.Request)
	authGroup.Post("/password-reset/confirm", middleware.AuthLinkRateLimit(rdb, cfg.RateRedis), passwordResetHandler.Confirm)
	apiV1.Post("/promotions/unsubscribe", middleware.AuthLinkRateLimit(rdb, cfg.RateRedis), unsubscribeHandler.Unsubscribe)

	passwordChanged := middleware.RequirePasswordChanged(sessionStore)
	selfWrite := middleware.SelfWriteRateLimit(rdb, cfg.RateRedis)

	apiV1.Get("/me", middleware.RequireAuth(tokenSigner), meHandler.Get)
	apiV1.Patch("/me", middleware.RequireAuthWithSession(tokenSigner, sessionStore), passwordChanged, selfWrite, meHandler.Patch)
	apiV1.Patch("/me/password", middleware.RequireAuthWithSession(tokenSigner, sessionStore), selfWrite, meHandler.PatchPassword)
	apiV1.Delete("/me", middleware.RequireAuthWithSession(tokenSigner, sessionStore), passwordChanged, selfWrite, accountErasureHandler.Erase)
	apiV1.Post(
		"/me/email-verification",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		passwordChanged,
		middleware.VerificationResendRateLimit(rdb, cfg.RateRedis),
		emailVerificationHandler.Resend,
	)
	apiV1.Get(
		"/me/export",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		passwordChanged,
		middleware.DataExportRateLimit(rdb, cfg.RateRedis),
		dataExportHandler.Export,
	)

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

	adminSettings := apiV1.Group(
		"/admin/settings",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleAdmin),
		passwordChanged,
	)
	adminSettings.Get("", adminStoreSettingsHandler.Get)
	adminSettings.Patch("", middleware.AdminWriteRateLimit(rdb, cfg.RateRedis), adminStoreSettingsHandler.Patch)

	adminClosedDates := apiV1.Group(
		"/admin/closed-dates",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleAdmin),
		passwordChanged,
	)
	adminClosedDates.Get("", adminStoreSettingsHandler.ListClosedDates)
	adminClosedDates.Post("", middleware.AdminWriteRateLimit(rdb, cfg.RateRedis), adminStoreSettingsHandler.CreateClosedDate)
	adminClosedDates.Delete("/:id", middleware.AdminWriteRateLimit(rdb, cfg.RateRedis), adminStoreSettingsHandler.DeleteClosedDate)

	// Public storefront read — proxy secret only; no JWT (guest browse).
	catalogRead := apiV1.Group("/catalog")
	catalogRead.Get("/categories", catalogHandler.ListCategories)
	catalogRead.Get("/products", catalogHandler.ListProducts)
	catalogRead.Get("/products/:id", catalogHandler.GetProduct)
	catalogRead.Get("/products/:id/reviews", reviewHandler.ProductReviews)
	catalogRead.Get("/combos", catalogHandler.ListCombos)
	catalogRead.Get("/combos/:id", catalogHandler.GetCombo)

	// Public: the checkout picker needs the pickup rules before anyone signs in.
	apiV1.Get("/store/pickup-rules", storeHandler.PickupRules)
	apiV1.Get("/store/pickup-slots", storeHandler.PickupSlots)

	customerCart := customerSessionGroup(apiV1, "/cart", tokenSigner, sessionStore, passwordChanged)
	customerCart.Get("", cartHandler.Get)
	customerCart.Post("/items", cartHandler.AddItem)
	customerCart.Patch("/items/:id", cartHandler.UpdateItem)
	customerCart.Delete("/items/:id", cartHandler.RemoveItem)
	customerCart.Put("/discount", middleware.DiscountAttemptRateLimit(rdb, cfg.RateRedis), cartHandler.ApplyDiscount)
	customerCart.Delete("/discount", cartHandler.RemoveDiscount)
	customerCart.Get("/reference-upload", middleware.ReferenceUploadRateLimit(rdb, cfg.RateRedis), mediaHandler.ReferenceImageSignature)

	customerOrders := customerSessionGroup(apiV1, "/orders", tokenSigner, sessionStore, passwordChanged)
	customerOrders.Post("/checkout", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), orderHandler.Checkout)
	customerOrders.Get("", orderHandler.List)
	customerOrders.Get("/:id", orderHandler.Get)
	customerOrders.Post("/:id/payment", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), paymentHandler.Start)
	customerOrders.Post("/:id/payment/capture", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), paymentHandler.Capture)
	customerOrders.Post("/:id/cancel", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), orderHandler.Cancel)
	customerOrders.Patch("/:id/pickup", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), orderHandler.Reschedule)
	customerOrders.Get("/:id/messages", conversationHandler.Thread)
	customerOrders.Post("/:id/messages", middleware.MessageWriteRateLimit(rdb, cfg.RateRedis), conversationHandler.Post)
	customerOrders.Post("/:id/messages/read", conversationHandler.MarkRead)
	customerOrders.Get("/:id/reviews", reviewHandler.ListByOrder)
	customerOrders.Post("/:id/reviews", middleware.ReviewWriteRateLimit(rdb, cfg.RateRedis), reviewHandler.Create)

	customerSaved := customerSessionGroup(apiV1, "/saved-products", tokenSigner, sessionStore, passwordChanged)
	customerSaved.Get("", savedProductHandler.List)
	customerSaved.Put("/:list/:product_id", savedProductHandler.Save)
	customerSaved.Delete("/:list/:product_id", savedProductHandler.Remove)
	// PayPal's notices: no session, each one signed and checked with PayPal.
	apiV1.Post("/payments/paypal/webhook", middleware.PaymentWebhookRateLimit(rdb, cfg.RateRedis), paymentHandler.Webhook)

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
	managerRead.Get("/reviews", managerReviewHandler.List)
	managerRead.Get("/reviews/summary", managerReviewHandler.Summary)
	managerRead.Get("/promotions", managerPromotionHandler.List)
	managerRead.Get("/promotions/audience", managerPromotionHandler.Audience)
	managerRead.Get("/incidents", managerIncidentHandler.List)
	managerRead.Get("/incidents/summary", managerIncidentHandler.Summary)
	managerRead.Get("/engagement", managerEngagementHandler.Report)
	managerRead.Get("/reports/sales", managerSalesReportHandler.Sales)
	managerRead.Get(
		"/media/cloudinary-signature",
		middleware.ManagerMediaRateLimit(rdb, cfg.RateRedis),
		mediaHandler.ProductImageSignature,
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
	managerWrite.Patch("/reviews/:id", managerReviewHandler.Moderate)
	managerWrite.Post("/promotions", managerPromotionHandler.Send)

	staffOrders := apiV1.Group(
		"/staff",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleStaff),
		passwordChanged,
	)
	staffOrders.Get("/orders", staffOrderHandler.List)
	staffOrders.Post("/orders", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), staffOrderHandler.Create)
	staffOrders.Post("/orders/quote", staffOrderHandler.Quote)
	staffOrders.Post("/customers/lookup", staffOrderHandler.FindCustomer)
	staffOrders.Get("/pickups", staffOrderHandler.Pickups)
	staffOrders.Get("/orders/:id", staffOrderHandler.Get)
	staffOrders.Patch("/orders/:id/status", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), staffOrderHandler.PatchStatus)
	staffOrders.Post("/orders/:id/incidents", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), staffOrderHandler.ReportIncident)
	staffOrders.Get("/tickets", staffTicketHandler.List)
	staffOrders.Patch("/tickets/:id/status", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), staffTicketHandler.PatchStatus)
	staffOrders.Patch("/tickets/:id/station", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), staffTicketHandler.Move)
	staffOrders.Get("/products", staffProductHandler.List)
	staffOrders.Patch("/products/:id/sold-out", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), staffProductHandler.PatchSoldOut)
	staffOrders.Get("/conversations", staffConversationHandler.List)
	staffOrders.Get("/conversations/counts", staffConversationHandler.Counts)
	staffOrders.Get("/orders/:id/messages", staffConversationHandler.Thread)
	staffOrders.Post("/orders/:id/messages", middleware.MessageWriteRateLimit(rdb, cfg.RateRedis), staffConversationHandler.Post)
	staffOrders.Post("/orders/:id/messages/read", staffConversationHandler.MarkRead)
	staffOrders.Patch("/orders/:id/conversation", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), staffConversationHandler.PatchStatus)

	bakerTickets := apiV1.Group(
		"/baker",
		middleware.RequireAuthWithSession(tokenSigner, sessionStore),
		middleware.RequireRole(domainuser.RoleBaker),
		passwordChanged,
	)
	bakerTickets.Get("/tickets", bakerTicketHandler.List)
	bakerTickets.Get("/tickets/:id", bakerTicketHandler.Get)
	bakerTickets.Patch("/tickets/:id/status", middleware.OrderWriteRateLimit(rdb, cfg.RateRedis), bakerTicketHandler.PatchStatus)

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
