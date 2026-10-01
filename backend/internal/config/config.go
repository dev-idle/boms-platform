package config

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

// Config holds all runtime configuration. Keep fields explicit—avoid map[string]any
// for application settings so validation stays compile-time friendly.
type Config struct {
	App        AppConfig
	HTTP       HTTPConfig
	CORS       CORSConfig
	Rate       RateLimitConfig
	RateRedis  RateLimitRedisConfig
	Postgres   PostgresConfig
	Redis      RedisConfig
	Outbox     OutboxConfig
	Order      OrderConfig
	Mail       MailConfig
	Realtime   RealtimeConfig
	Log        LogConfig
	JWT        JWTConfig
	Session    SessionConfig
	Cookie     CookieConfig
	Argon2     Argon2Config
	Seed       SeedConfig
	Cloudinary CloudinaryConfig
	PayPal     PayPalConfig
}

type AppConfig struct {
	Name  string
	Env   string
	Debug bool
	// SiteURL is the storefront origin: links in emails and PayPal's return
	// after a payment point to it.
	SiteURL string
}

type HTTPConfig struct {
	Host         string
	Port         int
	BodyLimit    int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	// RequestTimeout bounds one handler. Below WriteTimeout, so the handler
	// gives up before the socket does and the client sees a mapped error.
	RequestTimeout time.Duration
	TrustedProxies []string
	HSTSMaxAge     int    // seconds; 0 disables Strict-Transport-Security
	InternalSecret string // shared secret between Next.js proxy and this API; empty disables enforcement
}

type CORSConfig struct {
	AllowOrigins     []string
	AllowCredentials bool
	MaxAge           int
}

type RateLimitConfig struct {
	Max            int
	WindowDuration time.Duration
}

// RateLimitRedisConfig tunes Redis sliding-window limiters on auth and admin routes.
type RateLimitRedisConfig struct {
	AuthAttemptMax        int
	AuthAttemptWindow     time.Duration
	AuthRefreshMax        int
	AuthRefreshWindow     time.Duration
	AuthLogoutMax         int
	AuthLogoutWindow      time.Duration
	AdminWriteMax         int
	AdminWriteWindow      time.Duration
	ManagerWriteMax       int
	ManagerWriteWindow    time.Duration
	OrderWriteMax         int
	OrderWriteWindow      time.Duration
	SelfWriteMax          int
	SelfWriteWindow       time.Duration
	ManagerMediaMax       int
	ManagerMediaWindow    time.Duration
	ReferenceUploadMax    int
	ReferenceUploadWindow time.Duration
	AuthUserMax           int
	AuthUserWindow        time.Duration
	DiscountAttemptMax    int
	DiscountAttemptWindow time.Duration
	RealtimeTicketMax     int
	RealtimeTicketWindow  time.Duration
	DataExportMax         int
	DataExportWindow      time.Duration
	// Following an emailed link (confirming an address, setting a new password), per IP.
	AuthLinkMax    int
	AuthLinkWindow time.Duration
	// Asking for a password reset link, per IP.
	PasswordResetMax    int
	PasswordResetWindow time.Duration
	// Reset links sent to one account, whoever asks.
	PasswordResetAccountMax    int
	PasswordResetAccountWindow time.Duration
	// Asking for a new confirmation email, per user.
	VerificationResendMax    int
	VerificationResendWindow time.Duration
	// Payment provider notices, per IP.
	PaymentWebhookMax    int
	PaymentWebhookWindow time.Duration
}

type PostgresConfig struct {
	URL                string
	MaxConns           int32
	MinConns           int32
	MaxConnLifetime    time.Duration
	MaxConnIdleTime    time.Duration
	HealthCheckTimeout time.Duration
	// StatementTimeout bounds a single statement server-side. A request that is
	// abandoned by the client still holds its connection until the statement
	// ends, so without this one slow query can starve the pool.
	StatementTimeout time.Duration
}

type RedisConfig struct {
	Addr               string
	Password           string
	DB                 int
	PoolSize           int
	MinIdleConns       int
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	HealthCheckTimeout time.Duration
}

// RealtimeConfig tunes the push-only WebSocket listener. It runs on its own port
// so the API port stays private behind the BFF; browsers reach it with a
// single-use ticket the BFF obtained for them.
type RealtimeConfig struct {
	Addr                 string
	PublicURL            string
	AllowedOrigins       []string
	TicketTTL            time.Duration
	SessionCheckInterval time.Duration
	MaxLifetime          time.Duration
	PingInterval         time.Duration
	WriteTimeout         time.Duration
	// HandshakeTimeout bounds reading a handshake request and answering it.
	HandshakeTimeout time.Duration
	SendBuffer       int
	MaxConnsPerUser  int
	// MaxConns caps the connections one process holds, open sockets included.
	MaxConns int
	// AdmissionRate is how many handshakes per second the process redeems, so a
	// flood of forged tickets cannot starve the Redis pool the API shares.
	AdmissionRate int
	// AdmissionRatePerAddress is the share of it one client address may use, so
	// a single host cannot hold the whole budget.
	AdmissionRatePerAddress int
	// TrustedProxies are the edge proxies whose X-Forwarded-For names the client
	// address; any other peer is the client itself.
	TrustedProxies []netip.Prefix
}

// maxOutboxSweepBatch bounds one sweep transaction, which holds row locks while it publishes.
const maxOutboxSweepBatch = 1000

// OrderConfig tunes the worker's order jobs.
type OrderConfig struct {
	// JobInterval is how often the worker expires orders not paid in time,
	// makes the refunds cancelled orders ask for and records missed pickups.
	JobInterval time.Duration
}

// OutboxConfig tunes delivery of committed events. A post-commit delivery spends
// up to DispatchTimeout publishing and as long again recording the outcome, so
// SweepGrace must outlast twice DispatchTimeout: the sweeper then only picks up
// events that delivery has finished with.
type OutboxConfig struct {
	DispatchTimeout time.Duration
	SweepInterval   time.Duration
	SweepGrace      time.Duration
	SweepBatch      int32
	Retention       time.Duration
	PruneInterval   time.Duration
}

type LogConfig struct {
	Level            string
	Encoding         string
	EnableCaller     bool
	EnableStacktrace bool
}

type JWTConfig struct {
	Ed25519PrivateKey string // base64-encoded 32-byte seed (JWT_ED25519_PRIVATE_KEY)
	Issuer            string
	Audience          string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
}

// SessionConfig controls server-side session persistence in Redis.
type SessionConfig struct {
	TTL time.Duration
}

// CookieConfig holds HttpOnly cookie settings for future auth handlers.
type CookieConfig struct {
	Name   string
	Secure bool
	Domain string
	// RoleName carries the signed-in role so the Next.js proxy can route a
	// returning visitor at the edge. It is a navigation hint, never authority:
	// every route stays guarded by the session itself.
	RoleName string
}

// Argon2Config tunes the Argon2id password hasher (memory in KiB).
type Argon2Config struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type SeedConfig struct {
	DevAdminEmail    string
	DevAdminPassword string
	DevAdminFullName string
	DevAdminPhone    string
	DevAdminExtras   []DevAdminSeed
}

// CloudinaryConfig holds signed-upload credentials (API secret stays server-side only).
type CloudinaryConfig struct {
	CloudName    string
	APIKey       string
	APISecret    string
	UploadFolder string
	// ReferenceFolder holds customers' reference photos, one folder each.
	ReferenceFolder string
}

// Enabled reports whether signed product uploads can be issued.
func (c CloudinaryConfig) Enabled() bool {
	return strings.TrimSpace(c.CloudName) != "" &&
		strings.TrimSpace(c.APIKey) != "" &&
		strings.TrimSpace(c.APISecret) != ""
}

// ResolvedUploadFolder returns the trimmed upload folder with a safe default.
func (c CloudinaryConfig) ResolvedUploadFolder() string {
	folder := strings.TrimSpace(c.UploadFolder)
	if folder == "" {
		return "boms/products"
	}
	return strings.Trim(folder, "/")
}

// CustomerReferenceFolder is the folder a customer's reference photos go to: their
// own, under the configured base.
func (c CloudinaryConfig) CustomerReferenceFolder(userID uuid.UUID) string {
	base := strings.Trim(strings.TrimSpace(c.ReferenceFolder), "/")
	if base == "" {
		base = "boms/references"
	}
	return base + "/" + userID.String()
}

// Load reads the API's configuration from environment variables (defaults set
// with Viper; the command loads a local .env through godotenv before calling it)
// and validates all of it.
func Load() (*Config, error) {
	cfg, err := load()
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadWorker reads the configuration and validates only what cmd/worker uses.
func LoadWorker() (*Config, error) {
	cfg, err := load()
	if err != nil {
		return nil, err
	}
	if err := cfg.ValidateWorker(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func load() (*Config, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	maxConns, err := int32FromInt("postgres.max_conns", v.GetInt("postgres.max_conns"))
	if err != nil {
		return nil, err
	}
	minConns, err := int32FromInt("postgres.min_conns", v.GetInt("postgres.min_conns"))
	if err != nil {
		return nil, err
	}
	realtimeProxies, err := parsePrefixes("realtime.trusted_proxies", splitAndTrim(v.GetString("realtime.trusted_proxies")))
	if err != nil {
		return nil, err
	}
	sweepBatch, err := int32FromInt("outbox.sweep_batch", v.GetInt("outbox.sweep_batch"))
	if err != nil {
		return nil, err
	}

	argon2Memory, err := uint32FromInt("argon2.memory", v.GetInt("argon2.memory"))
	if err != nil {
		return nil, err
	}
	argon2Iterations, err := uint32FromInt("argon2.iterations", v.GetInt("argon2.iterations"))
	if err != nil {
		return nil, err
	}
	argon2Parallelism, err := uint8FromInt("argon2.parallelism", v.GetInt("argon2.parallelism"))
	if err != nil {
		return nil, err
	}
	argon2SaltLength, err := uint32FromInt("argon2.salt_length", v.GetInt("argon2.salt_length"))
	if err != nil {
		return nil, err
	}
	argon2KeyLength, err := uint32FromInt("argon2.key_length", v.GetInt("argon2.key_length"))
	if err != nil {
		return nil, err
	}

	devAdminExtras, err := parseDevAdminExtras(v.GetString("seed.dev_admin_extras"))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Name:    v.GetString("app.name"),
			Env:     v.GetString("app.env"),
			Debug:   v.GetBool("app.debug"),
			SiteURL: strings.TrimRight(strings.TrimSpace(v.GetString("app.site_url")), "/"),
		},
		HTTP: HTTPConfig{
			Host:           v.GetString("http.host"),
			Port:           v.GetInt("http.port"),
			BodyLimit:      v.GetInt("http.body_limit"),
			ReadTimeout:    v.GetDuration("http.read_timeout"),
			WriteTimeout:   v.GetDuration("http.write_timeout"),
			RequestTimeout: v.GetDuration("http.request_timeout"),
			IdleTimeout:    v.GetDuration("http.idle_timeout"),
			TrustedProxies: splitAndTrim(v.GetString("http.trusted_proxies")),
			HSTSMaxAge:     v.GetInt("http.hsts_max_age"),
			InternalSecret: strings.TrimSpace(v.GetString("http.internal_secret")),
		},
		CORS: CORSConfig{
			AllowOrigins:     splitAndTrim(v.GetString("cors.allow_origins")),
			AllowCredentials: v.GetBool("cors.allow_credentials"),
			MaxAge:           v.GetInt("cors.max_age"),
		},
		Rate: RateLimitConfig{
			Max:            v.GetInt("rate_limit.max"),
			WindowDuration: v.GetDuration("rate_limit.window"),
		},
		RateRedis: RateLimitRedisConfig{
			AuthAttemptMax:             v.GetInt("rate_limit.redis.auth_attempt_max"),
			AuthAttemptWindow:          v.GetDuration("rate_limit.redis.auth_attempt_window"),
			AuthRefreshMax:             v.GetInt("rate_limit.redis.auth_refresh_max"),
			AuthRefreshWindow:          v.GetDuration("rate_limit.redis.auth_refresh_window"),
			AuthLogoutMax:              v.GetInt("rate_limit.redis.auth_logout_max"),
			AuthLogoutWindow:           v.GetDuration("rate_limit.redis.auth_logout_window"),
			AdminWriteMax:              v.GetInt("rate_limit.redis.admin_write_max"),
			AdminWriteWindow:           v.GetDuration("rate_limit.redis.admin_write_window"),
			ManagerWriteMax:            v.GetInt("rate_limit.redis.manager_write_max"),
			ManagerWriteWindow:         v.GetDuration("rate_limit.redis.manager_write_window"),
			OrderWriteMax:              v.GetInt("rate_limit.redis.order_write_max"),
			OrderWriteWindow:           v.GetDuration("rate_limit.redis.order_write_window"),
			SelfWriteMax:               v.GetInt("rate_limit.redis.self_write_max"),
			SelfWriteWindow:            v.GetDuration("rate_limit.redis.self_write_window"),
			ManagerMediaMax:            v.GetInt("rate_limit.redis.manager_media_max"),
			ManagerMediaWindow:         v.GetDuration("rate_limit.redis.manager_media_window"),
			ReferenceUploadMax:         v.GetInt("rate_limit.redis.reference_upload_max"),
			ReferenceUploadWindow:      v.GetDuration("rate_limit.redis.reference_upload_window"),
			AuthUserMax:                v.GetInt("rate_limit.redis.auth_user_max"),
			AuthUserWindow:             v.GetDuration("rate_limit.redis.auth_user_window"),
			DiscountAttemptMax:         v.GetInt("rate_limit.redis.discount_attempt_max"),
			DiscountAttemptWindow:      v.GetDuration("rate_limit.redis.discount_attempt_window"),
			RealtimeTicketMax:          v.GetInt("rate_limit.redis.realtime_ticket_max"),
			RealtimeTicketWindow:       v.GetDuration("rate_limit.redis.realtime_ticket_window"),
			DataExportMax:              v.GetInt("rate_limit.redis.data_export_max"),
			DataExportWindow:           v.GetDuration("rate_limit.redis.data_export_window"),
			AuthLinkMax:                v.GetInt("rate_limit.redis.auth_link_max"),
			AuthLinkWindow:             v.GetDuration("rate_limit.redis.auth_link_window"),
			PasswordResetMax:           v.GetInt("rate_limit.redis.password_reset_max"),
			PasswordResetWindow:        v.GetDuration("rate_limit.redis.password_reset_window"),
			PasswordResetAccountMax:    v.GetInt("rate_limit.redis.password_reset_account_max"),
			PasswordResetAccountWindow: v.GetDuration("rate_limit.redis.password_reset_account_window"),
			VerificationResendMax:      v.GetInt("rate_limit.redis.verification_resend_max"),
			VerificationResendWindow:   v.GetDuration("rate_limit.redis.verification_resend_window"),
			PaymentWebhookMax:          v.GetInt("rate_limit.redis.payment_webhook_max"),
			PaymentWebhookWindow:       v.GetDuration("rate_limit.redis.payment_webhook_window"),
		},
		Postgres: PostgresConfig{
			URL:                v.GetString("postgres.url"),
			MaxConns:           maxConns,
			MinConns:           minConns,
			MaxConnLifetime:    v.GetDuration("postgres.max_conn_lifetime"),
			MaxConnIdleTime:    v.GetDuration("postgres.max_conn_idle_time"),
			HealthCheckTimeout: v.GetDuration("postgres.health_timeout"),
			StatementTimeout:   v.GetDuration("postgres.statement_timeout"),
		},
		Redis: RedisConfig{
			Addr:               v.GetString("redis.addr"),
			Password:           v.GetString("redis.password"),
			DB:                 v.GetInt("redis.db"),
			PoolSize:           v.GetInt("redis.pool_size"),
			MinIdleConns:       v.GetInt("redis.min_idle_conns"),
			DialTimeout:        v.GetDuration("redis.dial_timeout"),
			ReadTimeout:        v.GetDuration("redis.read_timeout"),
			WriteTimeout:       v.GetDuration("redis.write_timeout"),
			HealthCheckTimeout: v.GetDuration("redis.health_timeout"),
		},
		Order: OrderConfig{
			JobInterval: v.GetDuration("order.job_interval"),
		},
		Outbox: OutboxConfig{
			DispatchTimeout: v.GetDuration("outbox.dispatch_timeout"),
			SweepInterval:   v.GetDuration("outbox.sweep_interval"),
			SweepGrace:      v.GetDuration("outbox.sweep_grace"),
			SweepBatch:      sweepBatch,
			Retention:       v.GetDuration("outbox.retention"),
			PruneInterval:   v.GetDuration("outbox.prune_interval"),
		},
		Mail: MailConfig{
			SMTPHost:     v.GetString("mail.smtp_host"),
			SMTPPort:     v.GetInt("mail.smtp_port"),
			SMTPUsername: v.GetString("mail.smtp_username"),
			SMTPPassword: v.GetString("mail.smtp_password"),
			SMTPTLS:      strings.ToLower(strings.TrimSpace(v.GetString("mail.smtp_tls"))),
			FromAddress:  strings.TrimSpace(v.GetString("mail.from_address")),
			FromName:     strings.TrimSpace(v.GetString("mail.from_name")),
			ReplyTo:      strings.TrimSpace(v.GetString("mail.reply_to")),
			SendTimeout:  v.GetDuration("mail.send_timeout"),
			Concurrency:  v.GetInt("mail.concurrency"),
		},
		Realtime: RealtimeConfig{
			Addr:                    v.GetString("realtime.addr"),
			PublicURL:               v.GetString("realtime.public_url"),
			AllowedOrigins:          splitAndTrim(v.GetString("realtime.allowed_origins")),
			TicketTTL:               v.GetDuration("realtime.ticket_ttl"),
			SessionCheckInterval:    v.GetDuration("realtime.session_check_interval"),
			MaxLifetime:             v.GetDuration("realtime.max_lifetime"),
			PingInterval:            v.GetDuration("realtime.ping_interval"),
			WriteTimeout:            v.GetDuration("realtime.write_timeout"),
			HandshakeTimeout:        v.GetDuration("realtime.handshake_timeout"),
			SendBuffer:              v.GetInt("realtime.send_buffer"),
			MaxConnsPerUser:         v.GetInt("realtime.max_conns_per_user"),
			MaxConns:                v.GetInt("realtime.max_conns"),
			AdmissionRate:           v.GetInt("realtime.admission_rate"),
			AdmissionRatePerAddress: v.GetInt("realtime.admission_rate_per_address"),
			TrustedProxies:          realtimeProxies,
		},
		Log: LogConfig{
			Level:            v.GetString("log.level"),
			Encoding:         v.GetString("log.encoding"),
			EnableCaller:     v.GetBool("log.enable_caller"),
			EnableStacktrace: v.GetBool("log.enable_stacktrace"),
		},
		JWT: JWTConfig{
			Ed25519PrivateKey: v.GetString("jwt.ed25519_private_key"),
			Issuer:            v.GetString("jwt.issuer"),
			Audience:          v.GetString("jwt.audience"),
			AccessTTL:         v.GetDuration("jwt.access_ttl"),
			RefreshTTL:        v.GetDuration("jwt.refresh_ttl"),
		},
		Session: SessionConfig{
			TTL: v.GetDuration("session.ttl"),
		},
		Cookie: CookieConfig{
			Name:     v.GetString("cookie.name"),
			Secure:   v.GetBool("cookie.secure"),
			Domain:   v.GetString("cookie.domain"),
			RoleName: v.GetString("cookie.role_name"),
		},
		Argon2: Argon2Config{
			Memory:      argon2Memory,
			Iterations:  argon2Iterations,
			Parallelism: argon2Parallelism,
			SaltLength:  argon2SaltLength,
			KeyLength:   argon2KeyLength,
		},
		Seed: SeedConfig{
			DevAdminEmail:    v.GetString("seed.dev_admin_email"),
			DevAdminPassword: v.GetString("seed.dev_admin_password"),
			DevAdminFullName: v.GetString("seed.dev_admin_full_name"),
			DevAdminPhone:    v.GetString("seed.dev_admin_phone"),
			DevAdminExtras:   devAdminExtras,
		},
		Cloudinary: CloudinaryConfig{
			CloudName:       v.GetString("cloudinary.cloud_name"),
			APIKey:          v.GetString("cloudinary.api_key"),
			APISecret:       v.GetString("cloudinary.api_secret"),
			UploadFolder:    v.GetString("cloudinary.upload_folder"),
			ReferenceFolder: v.GetString("cloudinary.reference_folder"),
		},
		PayPal: PayPalConfig{
			Mode:         strings.TrimSpace(v.GetString("paypal.mode")),
			ClientID:     v.GetString("paypal.client_id"),
			ClientSecret: v.GetString("paypal.client_secret"),
			WebhookID:    strings.TrimSpace(v.GetString("paypal.webhook_id")),
		},
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "boms-api")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.debug", false)
	v.SetDefault("app.site_url", "http://localhost:3000")

	v.SetDefault("http.host", "127.0.0.1")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.body_limit", 1<<20)
	v.SetDefault("http.read_timeout", 15*time.Second)
	v.SetDefault("http.write_timeout", 30*time.Second)
	v.SetDefault("http.request_timeout", 20*time.Second)
	v.SetDefault("http.idle_timeout", 120*time.Second)
	v.SetDefault("http.hsts_max_age", 0)
	v.SetDefault("http.internal_secret", "")

	v.SetDefault("cors.allow_origins", "http://localhost:3000")
	v.SetDefault("cors.allow_credentials", false)
	v.SetDefault("cors.max_age", 86400)

	v.SetDefault("rate_limit.max", 120)
	v.SetDefault("rate_limit.window", 60*time.Second)

	v.SetDefault("rate_limit.redis.auth_attempt_max", 5)
	v.SetDefault("rate_limit.redis.auth_attempt_window", time.Minute)
	v.SetDefault("rate_limit.redis.auth_refresh_max", 10)
	v.SetDefault("rate_limit.redis.auth_refresh_window", time.Minute)
	v.SetDefault("rate_limit.redis.auth_logout_max", 10)
	v.SetDefault("rate_limit.redis.auth_logout_window", time.Minute)
	v.SetDefault("rate_limit.redis.admin_write_max", 30)
	v.SetDefault("rate_limit.redis.admin_write_window", time.Minute)
	v.SetDefault("rate_limit.redis.manager_write_max", 30)
	v.SetDefault("rate_limit.redis.manager_write_window", time.Minute)
	v.SetDefault("rate_limit.redis.order_write_max", 20)
	v.SetDefault("rate_limit.redis.order_write_window", time.Minute)
	v.SetDefault("rate_limit.redis.self_write_max", 10)
	v.SetDefault("rate_limit.redis.self_write_window", time.Minute)
	v.SetDefault("rate_limit.redis.manager_media_max", 20)
	v.SetDefault("rate_limit.redis.manager_media_window", time.Minute)
	v.SetDefault("rate_limit.redis.reference_upload_max", 10)
	v.SetDefault("rate_limit.redis.reference_upload_window", 10*time.Minute)
	v.SetDefault("rate_limit.redis.auth_user_max", 60)
	v.SetDefault("rate_limit.redis.auth_user_window", time.Minute)
	v.SetDefault("rate_limit.redis.discount_attempt_max", 10)
	v.SetDefault("rate_limit.redis.discount_attempt_window", 15*time.Minute)
	v.SetDefault("rate_limit.redis.realtime_ticket_max", 60)
	v.SetDefault("rate_limit.redis.realtime_ticket_window", time.Minute)
	v.SetDefault("rate_limit.redis.data_export_max", 5)
	v.SetDefault("rate_limit.redis.data_export_window", time.Hour)
	v.SetDefault("rate_limit.redis.auth_link_max", 10)
	v.SetDefault("rate_limit.redis.auth_link_window", 15*time.Minute)
	v.SetDefault("rate_limit.redis.password_reset_max", 5)
	v.SetDefault("rate_limit.redis.password_reset_window", 15*time.Minute)
	v.SetDefault("rate_limit.redis.password_reset_account_max", 3)
	v.SetDefault("rate_limit.redis.password_reset_account_window", time.Hour)
	v.SetDefault("rate_limit.redis.verification_resend_max", 3)
	v.SetDefault("rate_limit.redis.verification_resend_window", time.Hour)
	v.SetDefault("rate_limit.redis.payment_webhook_max", 60)
	v.SetDefault("rate_limit.redis.payment_webhook_window", time.Minute)

	// No default DB URL: use Neon (or any Postgres) via POSTGRES_URL in .env / environment.
	v.SetDefault("postgres.url", "")
	v.SetDefault("postgres.max_conns", 25)
	v.SetDefault("postgres.min_conns", 2)
	v.SetDefault("postgres.max_conn_lifetime", time.Hour)
	v.SetDefault("postgres.max_conn_idle_time", 15*time.Minute)
	v.SetDefault("postgres.health_timeout", 2*time.Second)
	v.SetDefault("postgres.statement_timeout", 10*time.Second)

	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 20)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.dial_timeout", 2*time.Second)
	v.SetDefault("redis.read_timeout", 2*time.Second)
	v.SetDefault("redis.write_timeout", 2*time.Second)
	v.SetDefault("redis.health_timeout", 2*time.Second)

	// Development sends to Mailpit (scripts/docker-compose.dev.yml), which keeps
	// every message and delivers none.
	v.SetDefault("mail.smtp_host", "127.0.0.1")
	v.SetDefault("mail.smtp_port", 1025)
	v.SetDefault("mail.smtp_tls", SMTPTLSNone)
	v.SetDefault("mail.from_address", "orders@chouxbakery.example")
	v.SetDefault("mail.from_name", "Choux")
	v.SetDefault("mail.reply_to", "hello@chouxbakery.example")
	v.SetDefault("mail.send_timeout", 10*time.Second)
	v.SetDefault("mail.concurrency", 4)

	v.SetDefault("outbox.dispatch_timeout", 5*time.Second)
	v.SetDefault("outbox.sweep_interval", 30*time.Second)
	v.SetDefault("outbox.sweep_grace", 15*time.Second)
	v.SetDefault("outbox.sweep_batch", 100)
	v.SetDefault("outbox.retention", 7*24*time.Hour)
	v.SetDefault("outbox.prune_interval", time.Hour)
	v.SetDefault("order.job_interval", time.Minute)

	v.SetDefault("realtime.addr", "127.0.0.1:8081")
	v.SetDefault("realtime.public_url", "ws://localhost:8081/ws")
	v.SetDefault("realtime.allowed_origins", "http://localhost:3000")
	v.SetDefault("realtime.ticket_ttl", 30*time.Second)
	v.SetDefault("realtime.session_check_interval", 30*time.Second)
	v.SetDefault("realtime.max_lifetime", time.Hour)
	v.SetDefault("realtime.ping_interval", 25*time.Second)
	v.SetDefault("realtime.write_timeout", 10*time.Second)
	v.SetDefault("realtime.handshake_timeout", 10*time.Second)
	v.SetDefault("realtime.send_buffer", 16)
	v.SetDefault("realtime.max_conns_per_user", 20)
	v.SetDefault("realtime.max_conns", 10000)
	v.SetDefault("realtime.admission_rate", 200)
	v.SetDefault("realtime.admission_rate_per_address", 10)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "json")
	v.SetDefault("log.enable_caller", false)
	v.SetDefault("log.enable_stacktrace", false)

	v.SetDefault("jwt.issuer", "boms-api")
	v.SetDefault("jwt.audience", "boms")
	v.SetDefault("jwt.access_ttl", 15*time.Minute)
	v.SetDefault("jwt.refresh_ttl", 168*time.Hour)

	v.SetDefault("session.ttl", 168*time.Hour)

	v.SetDefault("cookie.name", "boms_refresh")
	v.SetDefault("cookie.role_name", "boms_role")
	v.SetDefault("cookie.secure", false)
	v.SetDefault("cookie.domain", "")

	v.SetDefault("argon2.memory", 65536)
	v.SetDefault("argon2.iterations", 3)
	v.SetDefault("argon2.parallelism", 1)
	v.SetDefault("argon2.salt_length", 16)
	v.SetDefault("argon2.key_length", 32)

	v.SetDefault("seed.dev_admin_email", "")
	v.SetDefault("seed.dev_admin_password", "")
	v.SetDefault("seed.dev_admin_full_name", "Development Admin")
	v.SetDefault("seed.dev_admin_phone", "")
	v.SetDefault("seed.dev_admin_extras", "")

	v.SetDefault("cloudinary.cloud_name", "")
	v.SetDefault("cloudinary.api_key", "")
	v.SetDefault("cloudinary.api_secret", "")
	v.SetDefault("cloudinary.upload_folder", "boms/products")
	v.SetDefault("cloudinary.reference_folder", "boms/references")

	v.SetDefault("paypal.mode", PayPalModeSandbox)
	v.SetDefault("paypal.client_id", "")
	v.SetDefault("paypal.client_secret", "")
	v.SetDefault("paypal.webhook_id", "")
}

// Validate enforces production-safe constraints. Call after Load.
// ValidateWorker checks only what cmd/worker uses — the database, Redis, the
// outbox and order settings, PayPal and the mail server — so the worker never
// has to hold the API's signing key, internal secret or Cloudinary credentials.
func (c *Config) ValidateWorker() error {
	if err := c.validateStores(); err != nil {
		return err
	}
	if err := validateSiteURL(c.App.Env, c.App.SiteURL); err != nil {
		return err
	}
	if c.Order.JobInterval <= 0 {
		return errors.New("order.job_interval must be positive")
	}
	// Before an overdue order expires, the worker asks PayPal whether it was paid.
	if err := c.PayPal.validate(c.App.Env); err != nil {
		return err
	}
	return c.Mail.validate(c.App.Env)
}

// validateSiteURL checks the storefront origin links point to.
func validateSiteURL(env, raw string) error {
	site, err := url.Parse(raw)
	if err != nil || !isOrigin(site) {
		return errors.New("app.site_url must be an http:// or https:// origin: no path, query, fragment or credentials")
	}
	env = strings.ToLower(strings.TrimSpace(env))
	if (env == "production" || env == "staging") && site.Scheme != "https" {
		return errors.New("app.site_url must use https:// in staging/production")
	}
	return nil
}

func isOrigin(u *url.URL) bool {
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && u.User == nil && u.Opaque == "" &&
		strings.Trim(u.Path, "/") == "" && u.RawQuery == "" && !u.ForceQuery && u.Fragment == ""
}

// validateStores checks the settings every process that touches the database
// and the event bus shares.
func (c *Config) validateStores() error {
	if err := c.Outbox.validate(); err != nil {
		return err
	}
	if strings.TrimSpace(c.Postgres.URL) == "" {
		return errors.New("postgres.url is required")
	}
	if c.Postgres.MaxConns < 1 {
		return errors.New("postgres.max_conns must be at least 1")
	}
	if c.Postgres.MinConns < 0 || c.Postgres.MinConns > c.Postgres.MaxConns {
		return errors.New("postgres.min_conns must be between 0 and postgres.max_conns")
	}
	if strings.TrimSpace(c.Redis.Addr) == "" {
		return errors.New("redis.addr is required")
	}
	if c.Redis.PoolSize < 1 {
		return errors.New("redis.pool_size must be at least 1")
	}
	env := strings.ToLower(strings.TrimSpace(c.App.Env))
	if (env == "production" || env == "staging") && postgresTLSExplicitlyDisabled(c.Postgres.URL) {
		return errors.New("postgres.url must not disable TLS (sslmode=disable/allow) in staging/production")
	}
	return nil
}

func (c *Config) Validate() error {
	if c.HTTP.Port <= 0 || c.HTTP.Port > 65535 {
		return fmt.Errorf("http.port must be between 1 and 65535")
	}
	if c.HTTP.BodyLimit <= 0 {
		return errors.New("http.body_limit must be positive")
	}
	if c.HTTP.HSTSMaxAge < 0 {
		return errors.New("http.hsts_max_age must be >= 0")
	}
	const maxHSTS = 63072000 // 2 years (upper bound)
	if c.HTTP.HSTSMaxAge > maxHSTS {
		return fmt.Errorf("http.hsts_max_age must be <= %d seconds", maxHSTS)
	}
	if c.Rate.Max <= 0 {
		return errors.New("rate_limit.max must be positive")
	}
	if c.Rate.WindowDuration <= 0 {
		return errors.New("rate_limit.window must be positive")
	}
	if err := c.RateRedis.validate(); err != nil {
		return err
	}
	if err := c.validateStores(); err != nil {
		return err
	}
	if err := validateSiteURL(c.App.Env, c.App.SiteURL); err != nil {
		return err
	}
	if err := c.PayPal.validate(c.App.Env); err != nil {
		return err
	}
	if err := c.Realtime.validate(c.App.Env); err != nil {
		return err
	}
	if c.JWT.AccessTTL <= 0 {
		return errors.New("jwt.access_ttl must be positive")
	}
	if c.JWT.RefreshTTL <= 0 {
		return errors.New("jwt.refresh_ttl must be positive")
	}
	if c.Session.TTL <= 0 {
		return errors.New("session.ttl must be positive")
	}
	if c.Session.TTL < c.JWT.RefreshTTL {
		return errors.New("session.ttl must be >= jwt.refresh_ttl so server sessions cover refresh token lifetime")
	}
	if strings.TrimSpace(c.Cookie.Name) == "" {
		return errors.New("cookie.name must be set")
	}
	// An empty name makes fasthttp emit a nameless Set-Cookie the browser then
	// replays as a bare token; the same name as the session cookie overwrites the
	// refresh token with a role string and nobody can hold a session.
	if strings.TrimSpace(c.Cookie.RoleName) == "" {
		return errors.New("cookie.role_name must be set")
	}
	if c.Cookie.RoleName == c.Cookie.Name {
		return errors.New("cookie.role_name must differ from cookie.name")
	}
	// Fiber drops a cookie that net/http would reject without a word, so a bad
	// name or domain would ship a login that never sets its session cookie.
	for _, name := range []string{c.Cookie.Name, c.Cookie.RoleName} {
		probe := &http.Cookie{
			Name:     name,
			Value:    "v",
			Path:     "/",
			Domain:   c.Cookie.Domain,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		if err := probe.Valid(); err != nil {
			return fmt.Errorf("cookie settings would never reach the browser: %w", err)
		}
	}
	if slices.Contains(c.CORS.AllowOrigins, "*") && len(c.CORS.AllowOrigins) > 1 {
		return errors.New("cors.allow_origins: \"*\" allows every origin and cannot be listed with others")
	}
	if c.Argon2.Memory < 8*1024 {
		return errors.New("argon2.memory must be at least 8192 KiB")
	}
	if c.Argon2.Iterations < 1 {
		return errors.New("argon2.iterations must be at least 1")
	}
	if c.Argon2.Parallelism < 1 {
		return errors.New("argon2.parallelism must be at least 1")
	}
	if c.Argon2.SaltLength < 8 {
		return errors.New("argon2.salt_length must be at least 8")
	}
	if c.Argon2.KeyLength < 16 {
		return errors.New("argon2.key_length must be at least 16")
	}
	if strings.TrimSpace(c.Seed.DevAdminEmail) != "" && strings.TrimSpace(c.Seed.DevAdminPassword) == "" {
		return errors.New("seed.dev_admin_password is required when seed.dev_admin_email is set")
	}
	for _, admin := range c.Seed.DevAdmins() {
		if strings.TrimSpace(admin.Email) != "" && strings.TrimSpace(admin.Password) == "" {
			return errors.New("seed dev admin password is required when email is set")
		}
	}

	env := strings.ToLower(strings.TrimSpace(c.App.Env))
	if env == "production" || env == "staging" {
		if c.App.Debug {
			return fmt.Errorf("app.debug must be false when app.env is %q", c.App.Env)
		}
		if len(c.CORS.AllowOrigins) == 0 {
			return errors.New("cors.allow_origins must be set in non-development environments")
		}
		for _, o := range c.CORS.AllowOrigins {
			if o == "*" && c.CORS.AllowCredentials {
				return errors.New("cors: wildcard origin is incompatible with allow_credentials")
			}
		}
		if strings.TrimSpace(c.JWT.Ed25519PrivateKey) == "" {
			return errors.New("jwt.ed25519_private_key is required in non-development environments")
		}
		if c.HTTP.InternalSecret == "" {
			return errors.New("http.internal_secret is required in non-development environments (shared with the Next.js proxy)")
		}
		if len(c.HTTP.InternalSecret) < 32 {
			return errors.New("http.internal_secret must be at least 32 characters")
		}
		if strings.TrimSpace(c.JWT.Issuer) == "" {
			return errors.New("jwt.issuer is required in non-development environments")
		}
		if strings.TrimSpace(c.JWT.Audience) == "" {
			return errors.New("jwt.audience is required in non-development environments")
		}
		if !c.Cookie.Secure {
			return errors.New("cookie.secure must be true in non-development environments")
		}
		if c.HTTP.HSTSMaxAge <= 0 {
			return errors.New("http.hsts_max_age must be positive in non-development environments")
		}
		if !c.Cloudinary.Enabled() {
			return fmt.Errorf("cloudinary.cloud_name, cloudinary.api_key, and cloudinary.api_secret are required when app.env is %q", c.App.Env)
		}
	}

	return nil
}

// Realtime bounds that keep revocation prompt and a ticket short-lived. A
// ticket travels in a URL, so it must expire within the minute it was issued
// for; an open socket must notice an ended session within a minute.
const (
	maxRealtimeTicketTTL       = time.Minute
	maxRealtimeSessionInterval = time.Minute
)

func (c RealtimeConfig) validate(env string) error {
	env = strings.ToLower(strings.TrimSpace(env))
	deployed := env == "production" || env == "staging"
	if strings.TrimSpace(c.Addr) == "" {
		return errors.New("realtime.addr is required")
	}
	public, err := url.Parse(c.PublicURL)
	if err != nil || (public.Scheme != "ws" && public.Scheme != "wss") || public.Host == "" {
		return errors.New("realtime.public_url must be a ws:// or wss:// URL")
	}
	if deployed && public.Scheme != "wss" {
		return errors.New("realtime.public_url must use wss:// in non-development environments")
	}
	if err := validateRealtimeOrigins(c.AllowedOrigins, deployed); err != nil {
		return err
	}
	positive := []struct {
		name  string
		value time.Duration
	}{
		{"realtime.ticket_ttl", c.TicketTTL},
		{"realtime.session_check_interval", c.SessionCheckInterval},
		{"realtime.max_lifetime", c.MaxLifetime},
		{"realtime.ping_interval", c.PingInterval},
		{"realtime.write_timeout", c.WriteTimeout},
		{"realtime.handshake_timeout", c.HandshakeTimeout},
	}
	for _, d := range positive {
		if d.value <= 0 {
			return fmt.Errorf("%s must be positive", d.name)
		}
	}
	if c.TicketTTL > maxRealtimeTicketTTL {
		return fmt.Errorf("realtime.ticket_ttl must be at most %s", maxRealtimeTicketTTL)
	}
	if c.SessionCheckInterval > maxRealtimeSessionInterval {
		return fmt.Errorf("realtime.session_check_interval must be at most %s", maxRealtimeSessionInterval)
	}
	if c.SessionCheckInterval >= c.MaxLifetime {
		return errors.New("realtime.session_check_interval must be shorter than realtime.max_lifetime")
	}
	if c.SendBuffer < 1 {
		return errors.New("realtime.send_buffer must be at least 1")
	}
	if c.MaxConnsPerUser < 1 {
		return errors.New("realtime.max_conns_per_user must be at least 1")
	}
	if c.MaxConns < c.MaxConnsPerUser {
		return errors.New("realtime.max_conns must be at least realtime.max_conns_per_user")
	}
	if c.AdmissionRate < 1 {
		return errors.New("realtime.admission_rate must be at least 1")
	}
	if c.AdmissionRatePerAddress < 1 || c.AdmissionRatePerAddress > c.AdmissionRate {
		return errors.New("realtime.admission_rate_per_address must be between 1 and realtime.admission_rate")
	}
	return nil
}

// validateRealtimeOrigins accepts only origins exactly as a browser sends them —
// scheme://host[:port], lowercase, no default port — because the listener
// compares the Origin header byte for byte: anything else would silently refuse
// every handshake. Deployed sites are served over https from a real host.
func validateRealtimeOrigins(origins []string, deployed bool) error {
	if len(origins) == 0 {
		return errors.New("realtime.allowed_origins must list the site's origins")
	}
	for _, origin := range origins {
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" ||
			strings.ContainsAny(parsed.Host, "*?[]\\") || origin != parsed.Scheme+"://"+parsed.Host {
			return fmt.Errorf("realtime.allowed_origins: %q is not an exact http(s) origin", origin)
		}
		if origin != strings.ToLower(origin) || parsed.Port() == defaultPorts[parsed.Scheme] {
			return fmt.Errorf("realtime.allowed_origins: %q is not written as a browser sends it (lowercase, no default port)", origin)
		}
		if deployed && (parsed.Scheme != "https" || isLoopbackHost(parsed.Hostname())) {
			return fmt.Errorf("realtime.allowed_origins: %q must be an https origin on a public host in non-development environments", origin)
		}
	}
	return nil
}

var defaultPorts = map[string]string{"http": "80", "https": "443"}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (c OutboxConfig) validate() error {
	positive := []struct {
		name  string
		value time.Duration
	}{
		{"outbox.dispatch_timeout", c.DispatchTimeout},
		{"outbox.sweep_interval", c.SweepInterval},
		{"outbox.sweep_grace", c.SweepGrace},
		{"outbox.retention", c.Retention},
		{"outbox.prune_interval", c.PruneInterval},
	}
	for _, d := range positive {
		if d.value <= 0 {
			return fmt.Errorf("%s must be positive", d.name)
		}
	}
	if c.SweepBatch < 1 || c.SweepBatch > maxOutboxSweepBatch {
		return fmt.Errorf("outbox.sweep_batch must be between 1 and %d", maxOutboxSweepBatch)
	}
	if c.SweepGrace <= 2*c.DispatchTimeout {
		return errors.New("outbox.sweep_grace must be longer than twice outbox.dispatch_timeout, or the sweeper resends events still being delivered")
	}
	return nil
}

func (c RateLimitRedisConfig) validate() error {
	checks := []struct {
		name   string
		max    int
		window time.Duration
	}{
		{"rate_limit.redis.auth_attempt", c.AuthAttemptMax, c.AuthAttemptWindow},
		{"rate_limit.redis.auth_refresh", c.AuthRefreshMax, c.AuthRefreshWindow},
		{"rate_limit.redis.auth_logout", c.AuthLogoutMax, c.AuthLogoutWindow},
		{"rate_limit.redis.admin_write", c.AdminWriteMax, c.AdminWriteWindow},
		{"rate_limit.redis.manager_write", c.ManagerWriteMax, c.ManagerWriteWindow},
		{"rate_limit.redis.order_write", c.OrderWriteMax, c.OrderWriteWindow},
		{"rate_limit.redis.self_write", c.SelfWriteMax, c.SelfWriteWindow},
		{"rate_limit.redis.manager_media", c.ManagerMediaMax, c.ManagerMediaWindow},
		{"rate_limit.redis.reference_upload", c.ReferenceUploadMax, c.ReferenceUploadWindow},
		{"rate_limit.redis.auth_user", c.AuthUserMax, c.AuthUserWindow},
		{"rate_limit.redis.discount_attempt", c.DiscountAttemptMax, c.DiscountAttemptWindow},
		{"rate_limit.redis.realtime_ticket", c.RealtimeTicketMax, c.RealtimeTicketWindow},
		{"rate_limit.redis.data_export", c.DataExportMax, c.DataExportWindow},
		{"rate_limit.redis.auth_link", c.AuthLinkMax, c.AuthLinkWindow},
		{"rate_limit.redis.password_reset", c.PasswordResetMax, c.PasswordResetWindow},
		{"rate_limit.redis.password_reset_account", c.PasswordResetAccountMax, c.PasswordResetAccountWindow},
		{"rate_limit.redis.verification_resend", c.VerificationResendMax, c.VerificationResendWindow},
		{"rate_limit.redis.payment_webhook", c.PaymentWebhookMax, c.PaymentWebhookWindow},
	}
	for _, chk := range checks {
		if chk.max < 1 {
			return fmt.Errorf("%s_max must be positive", chk.name)
		}
		if chk.window <= 0 {
			return fmt.Errorf("%s_window must be positive", chk.name)
		}
	}
	return nil
}

func postgresTLSExplicitlyDisabled(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	mode := strings.ToLower(strings.TrimSpace(u.Query().Get("sslmode")))
	return mode == "disable" || mode == "allow"
}

func int32FromInt(field string, n int) (int32, error) {
	if n < math.MinInt32 || n > math.MaxInt32 {
		return 0, fmt.Errorf("%s: value %d out of int32 range", field, n)
	}
	return int32(n), nil
}

func uint32FromInt(field string, n int) (uint32, error) {
	if n < 0 || int64(n) > int64(^uint32(0)) {
		return 0, fmt.Errorf("%s: value %d out of uint32 range", field, n)
	}
	return uint32(n), nil
}

func uint8FromInt(field string, n int) (uint8, error) {
	if n < 0 || n > math.MaxUint8 {
		return 0, fmt.Errorf("%s: value %d out of uint8 range", field, n)
	}
	return uint8(n), nil
}

// parsePrefixes reads addresses and CIDR ranges; a bare address is its own range.
func parsePrefixes(field string, values []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		if prefix, err := netip.ParsePrefix(value); err == nil {
			prefixes = append(prefixes, prefix.Masked())
			continue
		}
		addr, err := netip.ParseAddr(value)
		if err != nil {
			return nil, fmt.Errorf("%s: %q is not an IP address or CIDR range", field, value)
		}
		addr = addr.Unmap()
		prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
	}
	return prefixes, nil
}

func splitAndTrim(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
