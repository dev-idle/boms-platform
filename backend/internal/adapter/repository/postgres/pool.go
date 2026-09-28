package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	"github.com/boms/backend/internal/config"
)

// Pool owns the pgx connection pool and the sqlc queries bound to it.
type Pool struct {
	inner   *pgxpool.Pool
	queries *sqlcgen.Queries
}

// NewPool creates a PostgreSQL pool and the sqlc queries that run on it.
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
	if cfg.StatementTimeout > 0 {
		// Server-side, so it still applies when the client goes away mid-query.
		pcfg.ConnConfig.RuntimeParams["statement_timeout"] =
			strconv.FormatInt(cfg.StatementTimeout.Milliseconds(), 10)
	}

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

	return &Pool{
		inner:   pool,
		queries: sqlcgen.New(pool),
	}, nil
}

// Queries exposes sqlc-generated methods (Querier); use in repository adapters.
func (p *Pool) Queries() *sqlcgen.Queries {
	if p == nil {
		return nil
	}
	return p.queries
}

// Close closes the pgx pool.
func (p *Pool) Close() {
	if p == nil {
		return
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

// Ping implements port.HealthResource.
func (p *Pool) Ping(ctx context.Context) error {
	if p == nil || p.inner == nil {
		return fmt.Errorf("postgres pool is nil")
	}
	return p.inner.Ping(ctx)
}

// rollbackTimeout bounds a rollback that runs after the caller's context may
// already be done — a request that timed out, or a panic unwinding.
const rollbackTimeout = 5 * time.Second

// WithTx runs fn in a transaction and commits on success. Repositories pick the
// transaction up from txCtx, so every call inside fn shares one connection.
func (p *Pool) WithTx(ctx context.Context, fn func(txCtx context.Context) error) (err error) {
	if p == nil || p.inner == nil {
		return fmt.Errorf("postgres pool is nil")
	}
	tx, err := p.inner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		// Runs on failure and on panic, so row locks and the pooled connection are
		// always given back; after a commit it is a no-op (ErrTxClosed). Detached
		// from ctx, which a timed-out request has already cancelled. A connection
		// that died mid-query was rolled back by the server when it closed.
		rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
		defer cancel()
		conn := tx.Conn()
		if rbErr := tx.Rollback(rbCtx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) && !conn.IsClosed() {
			err = errors.Join(err, fmt.Errorf("rollback tx: %w", rbErr))
		}
	}()
	if err = fn(withTx(ctx, tx)); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
