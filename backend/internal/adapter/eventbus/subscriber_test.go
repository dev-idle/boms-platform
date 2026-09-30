package eventbus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
)

type delivery struct {
	channel string
	payload string
}

// Everything the publisher sends reaches the subscriber, on the channel it was
// sent to, until the subscriber's context ends.
func TestSubscribe(t *testing.T) {
	t.Parallel()

	rdb := goredis.NewClient(&goredis.Options{Addr: miniredis.RunT(t).Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	var mu sync.Mutex
	var got []delivery
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		Subscribe(ctx, rdb, zap.NewNop(), func(channel string, payload []byte) {
			mu.Lock()
			defer mu.Unlock()
			got = append(got, delivery{channel: channel, payload: string(payload)})
		})
	}()
	require.Eventually(t, func() bool {
		n, err := rdb.PubSubNumPat(context.Background()).Result()
		return err == nil && n == 1
	}, 5*time.Second, 10*time.Millisecond, "the subscriber listens on one pattern")

	order := domainorder.Order{ID: uuid.New(), UserID: uuid.New(), Status: domainorder.StatusCancelled}
	event := domainorder.StatusChangedEvent(domainorder.StatusPending, order)
	require.NoError(t, NewRedisPublisher(rdb).Publish(context.Background(), []domainevent.Event{event}))

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(got) == 2
	}, 5*time.Second, 10*time.Millisecond)
	mu.Lock()
	channels := []string{got[0].channel, got[1].channel}
	payload := got[0].payload
	mu.Unlock()
	assert.ElementsMatch(t, []string{UserChannel(order.UserID), RoleChannel(domainuser.RoleStaff)}, channels)
	assert.Contains(t, payload, `"type":"order.status_changed"`)

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Subscribe must return once its context ends")
	}
}
