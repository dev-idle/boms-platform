package config_test

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/boms/backend/internal/config"
)

func base64Seed() string {
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i + 10)
	}
	return base64.StdEncoding.EncodeToString(seed)
}

func TestValidate_ProductionRequiresTLSWhenSSLModeSet(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		App: config.AppConfig{Env: "production", Debug: false},
		HTTP: config.HTTPConfig{
			Port:           8080,
			BodyLimit:      1024,
			ReadTimeout:    time.Second,
			WriteTimeout:   time.Second,
			IdleTimeout:    time.Second,
			InternalSecret: strings.Repeat("a", 32),
		},
		Rate:      config.RateLimitConfig{Max: 10, WindowDuration: time.Minute},
		RateRedis: defaultRateRedis(),
		Outbox:    defaultOutbox(),
		Realtime:  defaultRealtime(),
		Postgres: config.PostgresConfig{
			URL:                "postgres://host/db?sslmode=disable",
			MaxConns:           5,
			MinConns:           0,
			MaxConnLifetime:    time.Hour,
			MaxConnIdleTime:    time.Minute,
			HealthCheckTimeout: time.Second,
		},
		Redis: config.RedisConfig{
			Addr: "127.0.0.1:6379", PoolSize: 5,
			DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second,
			HealthCheckTimeout: time.Second,
		},
		CORS: config.CORSConfig{AllowOrigins: []string{"https://app.example.com"}},
		JWT: config.JWTConfig{
			Ed25519PrivateKey: base64Seed(),
			Issuer:            "boms-api",
			Audience:          "boms",
			AccessTTL:         time.Minute,
			RefreshTTL:        time.Hour,
		},
		Session: config.SessionConfig{TTL: time.Hour},
		Cookie:  config.CookieConfig{Name: "boms_refresh", RoleName: "boms_role", Secure: true},
		Argon2: config.Argon2Config{
			Memory: 65536, Iterations: 3, Parallelism: 1, SaltLength: 16, KeyLength: 32,
		},
	}
	cfg.HTTP.HSTSMaxAge = 31536000
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "TLS") {
		t.Fatalf("expected TLS validation error, got: %v", err)
	}
}

func TestValidate_DevelopmentAllowsSSLDisable(t *testing.T) {
	t.Parallel()
	cfg := minimalDevConfig()
	cfg.Postgres.URL = "postgres://host/db?sslmode=disable"
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_HSTSBounds(t *testing.T) {
	t.Parallel()
	cfg := minimalDevConfig()
	cfg.HTTP.HSTSMaxAge = -1
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "hsts") {
		t.Fatalf("expected hsts error, got %v", err)
	}
	cfg.HTTP.HSTSMaxAge = 63072001
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected hsts cap error")
	}
}

func TestValidate_ProductionRequiresCloudinary(t *testing.T) {
	t.Parallel()
	cfg := minimalProductionConfig()
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "cloudinary") {
		t.Fatalf("expected cloudinary validation error, got: %v", err)
	}
	cfg.Cloudinary = config.CloudinaryConfig{
		CloudName: "demo",
		APIKey:    "key",
		APISecret: "secret",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid production config with cloudinary: %v", err)
	}
}

func defaultRealtime() config.RealtimeConfig {
	return config.RealtimeConfig{
		Addr:                    "127.0.0.1:8081",
		PublicURL:               "ws://localhost:8081/ws",
		AllowedOrigins:          []string{"http://localhost:3000"},
		TicketTTL:               30 * time.Second,
		SessionCheckInterval:    30 * time.Second,
		MaxLifetime:             time.Hour,
		PingInterval:            25 * time.Second,
		WriteTimeout:            10 * time.Second,
		HandshakeTimeout:        10 * time.Second,
		SendBuffer:              16,
		MaxConnsPerUser:         20,
		MaxConns:                10000,
		AdmissionRate:           200,
		AdmissionRatePerAddress: 10,
	}
}

func defaultOutbox() config.OutboxConfig {
	return config.OutboxConfig{
		DispatchTimeout: 5 * time.Second,
		SweepInterval:   30 * time.Second,
		SweepGrace:      15 * time.Second,
		SweepBatch:      100,
		Retention:       7 * 24 * time.Hour,
		PruneInterval:   time.Hour,
	}
}

func defaultRateRedis() config.RateLimitRedisConfig {
	return config.RateLimitRedisConfig{
		AuthAttemptMax: 5, AuthAttemptWindow: time.Minute,
		AuthRefreshMax: 10, AuthRefreshWindow: time.Minute,
		AuthLogoutMax: 10, AuthLogoutWindow: time.Minute,
		AdminWriteMax: 30, AdminWriteWindow: time.Minute,
		ManagerWriteMax: 30, ManagerWriteWindow: time.Minute,
		OrderWriteMax: 20, OrderWriteWindow: time.Minute,
		SelfWriteMax: 10, SelfWriteWindow: time.Minute,
		ManagerMediaMax: 20, ManagerMediaWindow: time.Minute,
		AuthUserMax: 60, AuthUserWindow: time.Minute,
		DiscountAttemptMax: 10, DiscountAttemptWindow: 15 * time.Minute,
		RealtimeTicketMax: 60, RealtimeTicketWindow: time.Minute,
	}
}

func minimalProductionConfig() *config.Config {
	cfg := minimalDevConfig()
	cfg.App.Env = "production"
	cfg.CORS = config.CORSConfig{AllowOrigins: []string{"https://app.example.com"}}
	cfg.Realtime.PublicURL = "wss://app.example.com/ws"
	cfg.Realtime.AllowedOrigins = []string{"https://app.example.com"}
	cfg.HTTP.InternalSecret = strings.Repeat("a", 32)
	cfg.HTTP.HSTSMaxAge = 31536000
	cfg.Cookie.Secure = true
	cfg.JWT.Ed25519PrivateKey = base64Seed()
	cfg.JWT.Issuer = "boms-api"
	cfg.JWT.Audience = "boms"
	cfg.Postgres.URL = "postgres://host/db?sslmode=require"
	return cfg
}

func minimalDevConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{Env: "development", Debug: false},
		HTTP: config.HTTPConfig{
			Port: 8080, BodyLimit: 1024,
			ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second,
		},
		Rate:      config.RateLimitConfig{Max: 10, WindowDuration: time.Minute},
		RateRedis: defaultRateRedis(),
		Outbox:    defaultOutbox(),
		Realtime:  defaultRealtime(),
		Postgres: config.PostgresConfig{
			URL:      "postgres://host/db?sslmode=require",
			MaxConns: 5, MinConns: 0,
			MaxConnLifetime: time.Hour, MaxConnIdleTime: time.Minute, HealthCheckTimeout: time.Second,
		},
		Redis: config.RedisConfig{
			Addr: "127.0.0.1:6379", PoolSize: 5,
			DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second,
			HealthCheckTimeout: time.Second,
		},
		JWT: config.JWTConfig{
			AccessTTL:  time.Minute,
			RefreshTTL: time.Hour,
		},
		Session: config.SessionConfig{TTL: time.Hour},
		Cookie:  config.CookieConfig{Name: "boms_refresh", RoleName: "boms_role"},
		Argon2: config.Argon2Config{
			Memory: 65536, Iterations: 3, Parallelism: 1, SaltLength: 16, KeyLength: 32,
		},
	}
}

func TestValidate_RoleCookieName(t *testing.T) {
	t.Parallel()

	t.Run("must be set", func(t *testing.T) {
		t.Parallel()
		cfg := minimalDevConfig()
		cfg.Cookie.RoleName = ""
		if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "cookie.role_name must be set") {
			t.Fatalf("expected role cookie name error, got: %v", err)
		}
	})

	// Same name means the second Set-Cookie replaces the refresh token with a
	// role string, and nobody can hold a session.
	t.Run("must differ from the session cookie", func(t *testing.T) {
		t.Parallel()
		cfg := minimalDevConfig()
		cfg.Cookie.RoleName = cfg.Cookie.Name
		if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "must differ") {
			t.Fatalf("expected distinct cookie name error, got: %v", err)
		}
	})
}

// Fiber silently drops a cookie net/http would reject, so these must fail at
// startup instead of shipping a login that never sets its session cookie.
func TestValidate_CookieSettingsTheBrowserAccepts(t *testing.T) {
	t.Parallel()

	cases := map[string]func(*config.Config){
		"session_name_with_a_space":  func(c *config.Config) { c.Cookie.Name = "boms refresh" },
		"role_name_with_a_semicolon": func(c *config.Config) { c.Cookie.RoleName = "boms;role" },
		"domain_with_a_port":         func(c *config.Config) { c.Cookie.Domain = "choux.example:3000" },
	}
	for name, breakIt := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := minimalDevConfig()
			breakIt(cfg)
			if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "never reach the browser") {
				t.Fatalf("expected cookie settings error, got: %v", err)
			}
		})
	}

	t.Run("accepts_a_registrable_domain", func(t *testing.T) {
		t.Parallel()
		cfg := minimalDevConfig()
		cfg.Cookie.Domain = "choux.example"
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected valid config, got: %v", err)
		}
	})
}

func TestValidate_WildcardOriginStandsAlone(t *testing.T) {
	t.Parallel()
	cfg := minimalDevConfig()
	cfg.CORS.AllowOrigins = []string{"*", "https://choux.example"}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "cannot be listed with others") {
		t.Fatalf("expected wildcard origin error, got: %v", err)
	}
}

// The sweeper may only resend what the post-commit delivery gave up on.
func TestValidate_OutboxSweepOutlastsDelivery(t *testing.T) {
	t.Parallel()
	cfg := minimalDevConfig()
	cfg.Outbox.SweepGrace = 2 * cfg.Outbox.DispatchTimeout
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "sweep_grace must be longer than twice") {
		t.Fatalf("expected sweep grace error, got: %v", err)
	}
}

func TestValidate_OutboxBatch(t *testing.T) {
	t.Parallel()
	cfg := minimalDevConfig()
	cfg.Outbox.SweepBatch = 0
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "outbox.sweep_batch") {
		t.Fatalf("expected sweep batch error, got: %v", err)
	}
}

func TestValidate_OutboxBatchIsBounded(t *testing.T) {
	t.Parallel()
	cfg := minimalDevConfig()
	cfg.Outbox.SweepBatch = 1001
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "outbox.sweep_batch must be between") {
		t.Fatalf("expected sweep batch bound error, got: %v", err)
	}
}

// The worker never holds the API's signing key, internal secret or Cloudinary
// credentials, so its validation must not ask for them.
func TestValidateWorker_NeedsOnlyTheStores(t *testing.T) {
	t.Parallel()
	cfg := minimalProductionConfig()
	cfg.JWT.Ed25519PrivateKey = ""
	cfg.HTTP.InternalSecret = ""
	cfg.Cloudinary = config.CloudinaryConfig{}
	if err := cfg.ValidateWorker(); err != nil {
		t.Fatalf("expected worker config to be valid without API secrets, got: %v", err)
	}

	cfg.Redis.Addr = ""
	if err := cfg.ValidateWorker(); err == nil || !strings.Contains(err.Error(), "redis.addr") {
		t.Fatalf("expected redis error, got: %v", err)
	}
}

func TestValidate_Realtime(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		prod    bool
		breakIt func(*config.Config)
		want    string
	}{
		"plain_ws_in_production":   {prod: true, breakIt: func(c *config.Config) { c.Realtime.PublicURL = "ws://app.example.com/ws" }, want: "must use wss://"},
		"not_a_websocket_url":      {breakIt: func(c *config.Config) { c.Realtime.PublicURL = "http://localhost:8081/ws" }, want: "ws:// or wss://"},
		"no_allowed_origin":        {breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = nil }, want: "allowed_origins must list"},
		"wildcard_origin":          {breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"https://*.example.com"} }, want: "not an exact http(s) origin"},
		"origin_with_a_path":       {breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"http://localhost:3000/app"} }, want: "not an exact http(s) origin"},
		"origin_with_a_query":      {breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"https://app.example.com?*"} }, want: "not an exact http(s) origin"},
		"origin_with_a_slash":      {breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"http://localhost:3000/"} }, want: "not an exact http(s) origin"},
		"origin_with_userinfo":     {breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"http://user@localhost:3000"} }, want: "not an exact http(s) origin"},
		"plain_http_origin_live":   {prod: true, breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"http://app.example.com"} }, want: "must be an https origin"},
		"loopback_origin_live":     {prod: true, breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"https://localhost:3000"} }, want: "must be an https origin"},
		"loopback_ip_origin_live":  {prod: true, breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"https://127.0.0.1"} }, want: "must be an https origin"},
		"no_ticket_lifetime":       {breakIt: func(c *config.Config) { c.Realtime.TicketTTL = 0 }, want: "realtime.ticket_ttl"},
		"long_lived_ticket":        {breakIt: func(c *config.Config) { c.Realtime.TicketTTL = 24 * time.Hour }, want: "realtime.ticket_ttl must be at most"},
		"slow_revocation":          {breakIt: func(c *config.Config) { c.Realtime.SessionCheckInterval = time.Hour }, want: "realtime.session_check_interval must be at most"},
		"check_after_lifetime":     {breakIt: func(c *config.Config) { c.Realtime.MaxLifetime = 30 * time.Second }, want: "shorter than realtime.max_lifetime"},
		"no_handshake_timeout":     {breakIt: func(c *config.Config) { c.Realtime.HandshakeTimeout = 0 }, want: "realtime.handshake_timeout"},
		"no_room_to_queue_a_hint":  {breakIt: func(c *config.Config) { c.Realtime.SendBuffer = 0 }, want: "realtime.send_buffer"},
		"cap_below_per_user_cap":   {breakIt: func(c *config.Config) { c.Realtime.MaxConns = 5 }, want: "realtime.max_conns must be at least"},
		"no_admission_rate":        {breakIt: func(c *config.Config) { c.Realtime.AdmissionRate = 0 }, want: "realtime.admission_rate"},
		"one_address_takes_all":    {breakIt: func(c *config.Config) { c.Realtime.AdmissionRatePerAddress = 500 }, want: "realtime.admission_rate_per_address"},
		"no_share_per_address":     {breakIt: func(c *config.Config) { c.Realtime.AdmissionRatePerAddress = 0 }, want: "realtime.admission_rate_per_address"},
		"uppercase_origin":         {breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"http://Localhost:3000"} }, want: "as a browser sends it"},
		"origin_with_default_port": {prod: true, breakIt: func(c *config.Config) { c.Realtime.AllowedOrigins = []string{"https://app.example.com:443"} }, want: "as a browser sends it"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := minimalDevConfig()
			if tc.prod {
				cfg = minimalProductionConfig()
			}
			tc.breakIt(cfg)
			if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got: %v", tc.want, err)
			}
		})
	}
}
