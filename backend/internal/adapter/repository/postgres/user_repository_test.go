package postgres_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	"github.com/boms/backend/internal/config"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Parallel()

	ctx := context.Background()
	pgContainer, connStr, err := startPostgres(ctx, t)
	if err != nil {
		if strings.Contains(err.Error(), "docker") {
			t.Skip("docker not available")
		}
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	require.NoError(t, applyMigrations(ctx, connStr))

	pool, err := postgresadapter.NewPool(ctx, config.PostgresConfig{
		URL:                connStr,
		MaxConns:           5,
		MinConns:           1,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute,
		HealthCheckTimeout: 5 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	repo := postgresadapter.NewUserRepository(pool)

	t.Run("CRUD round trip", func(t *testing.T) {
		created, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "alice@example.com",
			PasswordHash: "$argon2id$v=19$m=65536,t=3,p=1$dummy",
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)
		assert.False(t, created.EmailVerified)

		byEmail, err := repo.GetByEmail(ctx, "alice@example.com")
		require.NoError(t, err)
		assert.Equal(t, created.ID, byEmail.ID)

		byID, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.Email, byID.Email)
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "alice@example.com",
			PasswordHash: "hash",
			Role:         domainuser.RoleCustomer,
		})
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrConflict.Code, ae.Code)
	})

	t.Run("soft deleted invisible", func(t *testing.T) {
		u, err := repo.Create(ctx, port.CreateUserParams{
			Email:        "ghost@example.com",
			PasswordHash: "hash",
			Role:         domainuser.RoleCustomer,
		})
		require.NoError(t, err)
		require.NoError(t, repo.SoftDelete(ctx, u.ID))
		_, err = repo.GetByID(ctx, u.ID)
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrNotFound.Code, ae.Code)
	})

	t.Run("GetByEmail not found", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nobody@example.com")
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrNotFound.Code, ae.Code)
	})
}

func startPostgres(ctx context.Context, t *testing.T) (*tcpostgres.PostgresContainer, string, error) {
	t.Helper()
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("boms_test"),
		tcpostgres.WithUsername("boms"),
		tcpostgres.WithPassword("boms"),
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

	dir := filepath.Join("..", "..", "..", "..", "migrations")
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
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			return err
		}
	}
	return nil
}
