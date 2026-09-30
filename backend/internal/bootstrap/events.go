package bootstrap

import (
	"github.com/redis/go-redis/v9"

	"github.com/boms/backend/internal/adapter/eventbus"
	"github.com/boms/backend/internal/adapter/queue"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/eventdispatch"
	"github.com/boms/backend/internal/service/notification"
)

// EventPublisher is every destination of a committed event: the realtime bus
// for open pages, then the queue of emails the event calls for. The API's
// post-commit delivery and the worker's sweeper both use it, so an event the
// sweeper recovers reaches exactly what it would have reached on time.
func EventPublisher(rdb *redis.Client) port.EventPublisher {
	return eventdispatch.Fanout(
		eventbus.NewRedisPublisher(rdb),
		notification.NewOrderEmails(queue.NewOrderEmailQueue(rdb)),
	)
}
