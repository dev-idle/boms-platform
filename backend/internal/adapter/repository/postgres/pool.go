package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	"github.com/boms/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// Pool wraps pgxpool for native ping/pgx usage, and sqlx + sqlc for type-safe SQL.
// sqlx sits on database/sql opened from the same pool (stdlib.OpenDBFromPool).
type Pool struct {
	inner   *pgxpool.Pool
	sqlxDB  *sqlx.DB
	queries *sqlcgen.Queries
}

// NewPool creates a PostgreSQL pool, a sqlx handle over it, and sqlc-generated queries.
func NewPool(ctx context.Context, cfg config.PostgresConfig) (*Pool, error) {
	pcfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres url: %w", err)
	}

	pcfg.MaxConns = cfg.MaxConns
	pcfg.MinConns = cfg.MinConns
	pcfg.MaxConnLifetime = cfg.MaxConnLifetime
	pcfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	pcfg.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.HealthCheckTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	sqlStd := stdlib.OpenDBFromPool(pool)
	sqlxDB := sqlx.NewDb(sqlStd, "pgx")

	return &Pool{
		inner:   pool,
		sqlxDB:  sqlxDB,
		queries: sqlcgen.New(sqlxDB),
	}, nil
}

// DB exposes the underlying pgx pool (raw SQL, COPY, LISTEN/NOTIFY, …).
func (p *Pool) DB() *pgxpool.Pool {
	if p == nil {
		return nil
	}
	return p.inner
}

// SQLX exposes sqlx for repositories that need NamedExec / StructScan helpers.
func (p *Pool) SQLX() *sqlx.DB {
	if p == nil {
		return nil
	}
	return p.sqlxDB
}

// Queries exposes sqlc-generated methods (Querier); use in repository adapters.
func (p *Pool) Queries() *sqlcgen.Queries {
	if p == nil {
		return nil
	}
	return p.queries
}

// Close closes sqlx (stdlib DB) then the pgx pool. Order matters for OpenDBFromPool.
func (p *Pool) Close() {
	if p == nil {
		return
	}
	if p.sqlxDB != nil {
		_ = p.sqlxDB.Close()
		p.sqlxDB = nil
	}
	if p.inner != nil {
		p.inner.Close()
		p.inner = nil
	}
	p.queries = nil
}

// Name implements port.HealthResource.
func (p *Pool) Name() string {
	return "postgres"
}

// Ping implements port.HealthResource (pgx pool; avoids sqlc/sqlx path for probes).
func (p *Pool) Ping(ctx context.Context) error {
	if p == nil || p.inner == nil {
		return fmt.Errorf("postgres pool is nil")
	}
	return p.inner.Ping(ctx)
}
