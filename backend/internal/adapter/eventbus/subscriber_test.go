package eventbus

import (
	"context"
	"sync"
	"sync/atomic"
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
		}, func() { t.Error("the first subscription has missed nothing") })
	}()
	require.Eventually(t, func() bool {
		n, err := rdb.PubSubNumPat(context.Background()).Result()
		return err == nil && n == 1
	}, 5*time.Second, 10*time.Millisecond, "the subscriber listens on one pattern")

	customer := uuid.New()
	order := domainorder.Order{ID: uuid.New(), UserID: &customer, Status: domainorder.StatusCancelled}
	event := domainorder.StatusChangedEvent(domainorder.StatusPending, order)
	require.NoError(t, NewRedisPublisher(rdb).Publish(context.Background(), []domainevent.Event{event}))

	want := []string{UserChannel(customer)}
	for _, role := range event.Audience.Roles {
		want = append(want, RoleChannel(role))
	}
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(got) == len(want)
	}, 5*time.Second, 10*time.Millisecond)
	mu.Lock()
	channels := make([]string, 0, len(got))
	for _, d := range got {
		channels = append(channels, d.channel)
	}
	payload := got[0].payload
	mu.Unlock()
	assert.ElementsMatch(t, want, channels)
	assert.Contains(t, payload, `"type":"order.status_changed"`)

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Subscribe must return once its context ends")
	}
}

// Events published while the subscription was down are lost: once it is made
// again, the sockets are told to read again — the first time it is made, not.
func TestSubscribe_TellsTheSocketsToReadAgainAfterADrop(t *testing.T) {
	t.Parallel()
	server := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: server.Addr(), MaxRetries: -1})
	t.Cleanup(func() { _ = rdb.Close() })
	var resyncs atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		Subscribe(ctx, rdb, zap.NewNop(), func(string, []byte) {}, func() { resyncs.Add(1) })
	}()
	subscribed := func() bool {
		n, err := rdb.PubSubNumPat(context.Background()).Result()
		return err == nil && n == 1
	}
	require.Eventually(t, subscribed, 5*time.Second, 10*time.Millisecond)
	assert.Zero(t, resyncs.Load())

	server.Close()
	require.NoError(t, server.Restart())

	require.Eventually(t, func() bool { return subscribed() && resyncs.Load() == 1 }, 10*time.Second, 20*time.Millisecond)
	cancel()
	<-done
}
