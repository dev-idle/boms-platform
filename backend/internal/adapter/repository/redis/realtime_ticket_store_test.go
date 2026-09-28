package redis_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	redisadapter "github.com/boms/backend/internal/adapter/repository/redis"
	"github.com/boms/backend/internal/config"
	domainrealtime "github.com/boms/backend/internal/domain/realtime"
	domainuser "github.com/boms/backend/internal/domain/user"
)

func newTestTicketStore(t *testing.T) (*redisadapter.RealtimeTicketStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client, err := redisadapter.NewClient(context.Background(), config.RedisConfig{
		Addr:               mr.Addr(),
		PoolSize:           4,
		DialTimeout:        time.Second,
		ReadTimeout:        time.Second,
		WriteTimeout:       time.Second,
		HealthCheckTimeout: time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	return redisadapter.NewRealtimeTicketStore(client), mr
}

// A ticket admits one socket for one session, once, within its lifetime.
func TestRealtimeTicketStore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ticket := domainrealtime.Ticket{UserID: uuid.New(), Role: domainuser.RoleStaff, SessionID: uuid.New()}

	t.Run("redeems_what_it_issued_exactly_once", func(t *testing.T) {
		t.Parallel()
		store, _ := newTestTicketStore(t)
		token, err := store.Issue(ctx, ticket, time.Minute)
		require.NoError(t, err)
		assert.Len(t, token, 43, "32 random bytes, base64url")

		got, err := store.Consume(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, ticket, got)

		_, err = store.Consume(ctx, token)
		assert.ErrorIs(t, err, domainrealtime.ErrTicketInvalid)
	})

	t.Run("refuses_an_expired_ticket", func(t *testing.T) {
		t.Parallel()
		store, mr := newTestTicketStore(t)
		token, err := store.Issue(ctx, ticket, 30*time.Second)
		require.NoError(t, err)
		mr.FastForward(31 * time.Second)

		_, err = store.Consume(ctx, token)
		assert.ErrorIs(t, err, domainrealtime.ErrTicketInvalid)
	})

	t.Run("refuses_a_malformed_or_unknown_token", func(t *testing.T) {
		t.Parallel()
		store, _ := newTestTicketStore(t)
		for _, token := range []string{"", "not-a-ticket", strings.Repeat("A", 43)} {
			_, err := store.Consume(ctx, token)
			assert.ErrorIs(t, err, domainrealtime.ErrTicketInvalid, token)
		}
	})

	t.Run("never_stores_the_token_itself", func(t *testing.T) {
		t.Parallel()
		store, mr := newTestTicketStore(t)
		token, err := store.Issue(ctx, ticket, time.Minute)
		require.NoError(t, err)
		for _, key := range mr.Keys() {
			assert.NotContains(t, key, token)
			value, _ := mr.Get(key)
			assert.False(t, strings.Contains(value, token))
		}
	})
}
