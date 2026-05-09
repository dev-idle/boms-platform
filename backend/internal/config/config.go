package config

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all runtime configuration. Keep fields explicit—avoid map[string]any
// for application settings so validation stays compile-time friendly.
type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	CORS     CORSConfig
	Rate     RateLimitConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Asynq    AsynqConfig
	Log      LogConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Name  string
	Env   string
	Debug bool
}

type HTTPConfig struct {
	Host           string
	Port           int
	BodyLimit      int
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	TrustedProxies []string
	HSTSMaxAge     int // seconds; 0 disables Strict-Transport-Security
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

type PostgresConfig struct {
	URL                string
	MaxConns           int32
	MinConns           int32
	MaxConnLifetime    time.Duration
	MaxConnIdleTime    time.Duration
	HealthCheckTimeout time.Duration
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

type AsynqConfig struct {
	Enabled bool
}

type LogConfig struct {
	Level            string
	Encoding         string
	EnableCaller     bool
	EnableStacktrace bool
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// Load reads configuration from environment variables (set defaults with Viper;
// load a local .env into the environment from cmd/api via godotenv before Load).
func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	cfg := &Config{
		App: AppConfig{
			Name:  v.GetString("app.name"),
			Env:   v.GetString("app.env"),
			Debug: v.GetBool("app.debug"),
		},
		HTTP: HTTPConfig{
			Host:           v.GetString("http.host"),
			Port:           v.GetInt("http.port"),
			BodyLimit:      v.GetInt("http.body_limit"),
			ReadTimeout:    v.GetDuration("http.read_timeout"),
			WriteTimeout:   v.GetDuration("http.write_timeout"),
			IdleTimeout:    v.GetDuration("http.idle_timeout"),
			TrustedProxies: splitAndTrim(v.GetString("http.trusted_proxies")),
			HSTSMaxAge:     v.GetInt("http.hsts_max_age"),
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
		Postgres: PostgresConfig{
			URL:                v.GetString("postgres.url"),
			MaxConns:           int32(v.GetInt("postgres.max_conns")),
			MinConns:           int32(v.GetInt("postgres.min_conns")),
			MaxConnLifetime:    v.GetDuration("postgres.max_conn_lifetime"),
			MaxConnIdleTime:    v.GetDuration("postgres.max_conn_idle_time"),
			HealthCheckTimeout: v.GetDuration("postgres.health_timeout"),
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
		Asynq: AsynqConfig{
			Enabled: v.GetBool("asynq.enabled"),
		},
		Log: LogConfig{
			Level:            v.GetString("log.level"),
			Encoding:         v.GetString("log.encoding"),
			EnableCaller:     v.GetBool("log.enable_caller"),
			EnableStacktrace: v.GetBool("log.enable_stacktrace"),
		},
		JWT: JWTConfig{
			AccessSecret:  v.GetString("jwt.access_secret"),
			RefreshSecret: v.GetString("jwt.refresh_secret"),
			AccessTTL:     v.GetDuration("jwt.access_ttl"),
			RefreshTTL:    v.GetDuration("jwt.refresh_ttl"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "boms-api")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.debug", false)

	v.SetDefault("http.host", "127.0.0.1")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.body_limit", 1<<20)
	v.SetDefault("http.read_timeout", 15*time.Second)
	v.SetDefault("http.write_timeout", 30*time.Second)
	v.SetDefault("http.idle_timeout", 120*time.Second)
	v.SetDefault("http.hsts_max_age", 0)

	v.SetDefault("cors.allow_origins", "http://localhost:3000")
	v.SetDefault("cors.allow_credentials", false)
	v.SetDefault("cors.max_age", 86400)

	v.SetDefault("rate_limit.max", 120)
	v.SetDefault("rate_limit.window", 60*time.Second)

	// No default DB URL: use Neon (or any Postgres) via POSTGRES_URL in .env / environment.
	v.SetDefault("postgres.url", "")
	v.SetDefault("postgres.max_conns", 25)
	v.SetDefault("postgres.min_conns", 2)
	v.SetDefault("postgres.max_conn_lifetime", time.Hour)
	v.SetDefault("postgres.max_conn_idle_time", 15*time.Minute)
	v.SetDefault("postgres.health_timeout", 2*time.Second)

	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 20)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.dial_timeout", 2*time.Second)
	v.SetDefault("redis.read_timeout", 2*time.Second)
	v.SetDefault("redis.write_timeout", 2*time.Second)
	v.SetDefault("redis.health_timeout", 2*time.Second)

	v.SetDefault("asynq.enabled", false)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "json")
	v.SetDefault("log.enable_caller", false)
	v.SetDefault("log.enable_stacktrace", false)

	v.SetDefault("jwt.access_ttl", 15*time.Minute)
	v.SetDefault("jwt.refresh_ttl", 168*time.Hour)
}

// Validate enforces production-safe constraints. Call after Load.
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
		if len(c.JWT.AccessSecret) < 32 {
			return errors.New("jwt.access_secret must be at least 32 characters in non-development environments")
		}
		if len(c.JWT.RefreshSecret) < 32 {
			return errors.New("jwt.refresh_secret must be at least 32 characters in non-development environments")
		}
		if postgresTLSExplicitlyDisabled(c.Postgres.URL) {
			return errors.New("postgres.url must not disable TLS (sslmode=disable/allow) in staging/production")
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
