package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redisadapter "github.com/boms/backend/internal/adapter/repository/redis"
	"github.com/boms/backend/internal/config"
	domainsession "github.com/boms/backend/internal/domain/session"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSessionTTL = 7 * 24 * time.Hour

func newTestSessionStore(t *testing.T) (*redisadapter.SessionStore, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client, err := redisadapter.NewClient(context.Background(), config.RedisConfig{
		Addr:               mr.Addr(),
		PoolSize:           4,
		MinIdleConns:       1,
		DialTimeout:        time.Second,
		ReadTimeout:        time.Second,
		WriteTimeout:       time.Second,
		HealthCheckTimeout: time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	return redisadapter.NewSessionStore(client, testSessionTTL), mr
}

func sampleMeta() domainsession.SessionMeta {
	return domainsession.SessionMeta{
		RefreshJTI: "jti-1",
		CreatedAt:  time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		UserAgent:  "test-agent",
		IP:         "127.0.0.1",
	}
}

func TestSessionStore(t *testing.T) {
	t.Parallel()

	t.Run("CRUD", func(t *testing.T) {
		t.Parallel()
		store, _ := newTestSessionStore(t)
		ctx := context.Background()
		meta := sampleMeta()

		require.NoError(t, store.Create(ctx, "user-1", "sess-1", meta))
		got, err := store.Get(ctx, "user-1", "sess-1")
		require.NoError(t, err)
		assert.Equal(t, meta.RefreshJTI, got.RefreshJTI)
		assert.Equal(t, meta.UserAgent, got.UserAgent)

		require.NoError(t, store.Delete(ctx, "user-1", "sess-1"))
		_, err = store.Get(ctx, "user-1", "sess-1")
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrNotFound.Code, ae.Code)
	})

	t.Run("DeleteAllForUser scope", func(t *testing.T) {
		t.Parallel()
		store, mr := newTestSessionStore(t)
		ctx := context.Background()
		meta := sampleMeta()
		require.NoError(t, store.Create(ctx, "user-9", "s1", meta))
		require.NoError(t, store.Create(ctx, "user-9", "s2", meta))
		require.NoError(t, store.Create(ctx, "user-9", "s3", meta))
		require.NoError(t, store.Create(ctx, "user-other", "s1", meta))

		require.NoError(t, store.DeleteAllForUser(ctx, "user-9"))
		assert.False(t, mr.Exists("session:user-9:s1"))
		assert.False(t, mr.Exists("session:user-9:s2"))
		assert.True(t, mr.Exists("session:user-other:s1"))
	})

	t.Run("Rotate", func(t *testing.T) {
		t.Parallel()
		store, mr := newTestSessionStore(t)
		ctx := context.Background()
		meta := sampleMeta()
		require.NoError(t, store.Create(ctx, "user-1", "old-sess", meta))

		newMeta := sampleMeta()
		newMeta.RefreshJTI = "jti-rotated"
		require.NoError(t, store.Rotate(ctx, "user-1", "old-sess", "new-sess", meta.RefreshJTI, newMeta))

		assert.False(t, mr.Exists("session:user-1:old-sess"))
		assert.True(t, mr.Exists("session:user-1:new-sess"))
		got, err := store.Get(ctx, "user-1", "new-sess")
		require.NoError(t, err)
		assert.Equal(t, "jti-rotated", got.RefreshJTI)
	})

	t.Run("Rotate JTI mismatch", func(t *testing.T) {
		t.Parallel()
		store, _ := newTestSessionStore(t)
		ctx := context.Background()
		meta := sampleMeta()
		require.NoError(t, store.Create(ctx, "user-1", "old-sess", meta))

		newMeta := sampleMeta()
		newMeta.RefreshJTI = "jti-new"
		err := store.Rotate(ctx, "user-1", "old-sess", "new-sess", "wrong-jti", newMeta)
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrConflict.Code, ae.Code)
	})

	t.Run("Rotate missing old session", func(t *testing.T) {
		t.Parallel()
		store, _ := newTestSessionStore(t)
		ctx := context.Background()
		err := store.Rotate(ctx, "user-1", "gone", "new-sess", "jti-1", sampleMeta())
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrNotFound.Code, ae.Code)
	})

	t.Run("TTL", func(t *testing.T) {
		t.Parallel()
		store, mr := newTestSessionStore(t)
		ctx := context.Background()
		require.NoError(t, store.Create(ctx, "user-1", "sess-ttl", sampleMeta()))
		mr.FastForward(8 * 24 * time.Hour)
		_, err := store.Get(ctx, "user-1", "sess-ttl")
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrNotFound.Code, ae.Code)
	})
}
