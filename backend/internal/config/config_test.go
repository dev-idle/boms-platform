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
			Port:        8080,
			BodyLimit:   1024,
			ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second,
		},
		Rate: config.RateLimitConfig{Max: 10, WindowDuration: time.Minute},
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
		Cookie:  config.CookieConfig{Name: "boms_refresh"},
		Argon2: config.Argon2Config{
			Memory: 65536, Iterations: 3, Parallelism: 1, SaltLength: 16, KeyLength: 32,
		},
	}
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

func minimalDevConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{Env: "development", Debug: false},
		HTTP: config.HTTPConfig{
			Port: 8080, BodyLimit: 1024,
			ReadTimeout: time.Second, WriteTimeout: time.Second, IdleTimeout: time.Second,
		},
		Rate: config.RateLimitConfig{Max: 10, WindowDuration: time.Minute},
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
		Cookie:  config.CookieConfig{Name: "boms_refresh"},
		Argon2: config.Argon2Config{
			Memory: 65536, Iterations: 3, Parallelism: 1, SaltLength: 16, KeyLength: 32,
		},
	}
}
