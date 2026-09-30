package postgres_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	"github.com/boms/backend/internal/config"
)

// Test fixtures — not real credentials (gosec G101).
const testPasswordHashFixture = "test-password-hash-fixture"

// newIntegrationPool starts a throwaway Postgres, applies every migration and
// returns a pool on it that closes with the test. The test is skipped in short
// mode and when Docker is unavailable.
func newIntegrationPool(t *testing.T, maxConns int32) *postgresadapter.Pool {
	t.Helper()
	pool, _ := newIntegrationDB(t, maxConns)
	return pool
}

// newIntegrationDB is newIntegrationPool plus the connection string, for a
// test that must set up a state no repository writes (an expired row, say).
func newIntegrationDB(t *testing.T, maxConns int32) (*postgresadapter.Pool, string) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	container, connStr, err := startPostgres(ctx, t)
	if err != nil {
		if strings.Contains(err.Error(), "docker") {
			t.Skip("docker not available")
		}
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })
	require.NoError(t, applyMigrations(ctx, connStr))

	pool, err := postgresadapter.NewPool(ctx, config.PostgresConfig{
		URL:                connStr,
		MaxConns:           maxConns,
		MinConns:           1,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute,
		HealthCheckTimeout: 5 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool, connStr
}

func startPostgres(ctx context.Context, t *testing.T) (*tcpostgres.PostgresContainer, string, error) {
	t.Helper()
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("boms_test"),
		tcpostgres.WithUsername("boms"),
		tcpostgres.WithPassword("boms"), //nolint:gosec // G101: ephemeral testcontainer password
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		return nil, "", err
	}
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	return container, connStr, err
}

func applyMigrations(ctx context.Context, connStr string) error {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return err
	}
	defer pool.Close()

	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "..", "migrations"))
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	for _, name := range files {
		body, err := readMigrationSQL(dir, name)
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			return err
		}
	}
	return nil
}

func readMigrationSQL(baseDir, name string) ([]byte, error) {
	if name != filepath.Base(name) || strings.Contains(name, "..") {
		return nil, fmt.Errorf("invalid migration filename: %q", name)
	}
	path := filepath.Join(baseDir, name)
	clean := filepath.Clean(path)
	rel, err := filepath.Rel(baseDir, clean)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("migration path outside base: %q", name)
	}
	return os.ReadFile(clean) //nolint:gosec // G304: path constrained to migrations/ basenames from ReadDir
}
