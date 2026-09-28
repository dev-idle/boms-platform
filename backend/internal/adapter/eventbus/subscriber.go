package eventbus

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// resubscribeDelay spaces out attempts to subscribe again after the bus refused.
const resubscribeDelay = time.Second

// Subscribe hands every event published on the bus to deliver, with the channel
// it came on, until ctx ends. One subscription serves the whole process however
// many sockets it holds; go-redis restores it after a dropped connection, and a
// refused subscription is retried.
func Subscribe(ctx context.Context, rdb *goredis.Client, log *zap.Logger, deliver func(channel string, payload []byte)) {
	for ctx.Err() == nil {
		if err := subscribeOnce(ctx, rdb, deliver); err != nil {
			log.Warn("event_bus_subscribe_failed", zap.Error(err))
			select {
			case <-ctx.Done():
			case <-time.After(resubscribeDelay):
			}
		}
	}
}

func subscribeOnce(ctx context.Context, rdb *goredis.Client, deliver func(channel string, payload []byte)) error {
	sub := rdb.PSubscribe(ctx, channelPrefix+"*")
	defer func() { _ = sub.Close() }()
	if _, err := sub.Receive(ctx); err != nil {
		return fmt.Errorf("subscribe to events: %w", err)
	}
	messages := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-messages:
			if !ok {
				return fmt.Errorf("event subscription closed")
			}
			deliver(msg.Channel, []byte(msg.Payload))
		}
	}
}
