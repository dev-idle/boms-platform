package eventbus

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
)

func TestRedisPublisher_Publish(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	rdb := goredis.NewClient(&goredis.Options{Addr: miniredis.RunT(t).Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	order := domainorder.Order{ID: uuid.New(), UserID: uuid.New(), Status: domainorder.StatusReady}
	event := domainorder.StatusChangedEvent(domainorder.StatusInProduction, order)
	event.OccurredAt = time.Date(2026, 9, 28, 9, 0, 0, 0, time.UTC)

	sub := rdb.Subscribe(ctx,
		userChannel(order.UserID),
		roleChannel(domainuser.RoleStaff),
		roleChannel(domainuser.RoleBaker),
		userChannel(uuid.New()),
	)
	t.Cleanup(func() { _ = sub.Close() })
	_, err := sub.Receive(ctx)
	require.NoError(t, err)

	require.NoError(t, NewRedisPublisher(rdb).Publish(ctx, []domainevent.Event{event}))

	got := map[string]message{}
	for range 3 {
		msg, err := sub.ReceiveMessage(ctx)
		require.NoError(t, err)
		var decoded message
		require.NoError(t, json.Unmarshal([]byte(msg.Payload), &decoded))
		got[msg.Channel] = decoded
	}

	want := message{
		ID:   event.ID.String(),
		Type: "order.status_changed",
		At:   event.OccurredAt,
		Data: map[string]string{"order_id": order.ID.String(), "status": "ready"},
	}
	assert.Equal(t, map[string]message{
		userChannel(order.UserID):         want,
		roleChannel(domainuser.RoleStaff): want,
		roleChannel(domainuser.RoleBaker): want,
	}, got, "only the order's customer and the order roles hear about it")
}

func TestRedisPublisher_ReportsAnUnreachableBus(t *testing.T) {
	t.Parallel()

	server := miniredis.RunT(t)
	rdb := goredis.NewClient(&goredis.Options{Addr: server.Addr(), MaxRetries: -1})
	t.Cleanup(func() { _ = rdb.Close() })
	server.Close()

	order := domainorder.Order{ID: uuid.New(), UserID: uuid.New(), Status: domainorder.StatusPending}
	err := NewRedisPublisher(rdb).Publish(context.Background(), []domainevent.Event{domainorder.CreatedEvent(order)})
	assert.Error(t, err)
}
