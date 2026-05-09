package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/boms/backend/internal/config"
)

func TestValidate_ProductionRequiresTLSWhenSSLModeSet(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{
		App: config.AppConfig{Env: "production", Debug: false},
		HTTP: config.HTTPConfig{
			Port:        8080,
			BodyLimit:   1024,
			ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second,
		},
		Rate: config.RateLimitConfig{Max: 10, WindowDuration: time.Minute},
		Postgres: config.PostgresConfig{
			URL:                "postgres://u:p@host/db?sslmode=disable",
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
			AccessSecret:  strings.Repeat("a", 32),
			RefreshSecret: strings.Repeat("b", 32),
			AccessTTL:     time.Minute,
			RefreshTTL:    time.Hour,
		},
	}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "TLS") {
		t.Fatalf("expected TLS validation error, got: %v", err)
	}
}

func TestValidate_DevelopmentAllowsSSLDisable(t *testing.T) {
	t.Parallel()
	cfg := minimalDevConfig()
	cfg.Postgres.URL = "postgres://u:p@host/db?sslmode=disable"
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

func minimalDevConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{Env: "development", Debug: false},
		HTTP: config.HTTPConfig{
			Port: 8080, BodyLimit: 1024,
			ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second,
		},
		Rate: config.RateLimitConfig{Max: 10, WindowDuration: time.Minute},
		Postgres: config.PostgresConfig{
			URL:      "postgres://u:p@host/db?sslmode=require",
			MaxConns: 5, MinConns: 0,
			MaxConnLifetime: time.Hour, MaxConnIdleTime: time.Minute, HealthCheckTimeout: time.Second,
		},
		Redis: config.RedisConfig{
			Addr: "127.0.0.1:6379", PoolSize: 5,
			DialTimeout: time.Second, ReadTimeout: time.Second, WriteTimeout: time.Second,
			HealthCheckTimeout: time.Second,
		},
	}
}
