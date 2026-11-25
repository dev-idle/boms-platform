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
// many sockets it holds; a dropped or refused one is made again. Events
// published while it was down are lost, so every subscription after the first
// calls resync, which tells the sockets to read again.
func Subscribe(
	ctx context.Context,
	rdb *goredis.Client,
	log *zap.Logger,
	deliver func(channel string, payload []byte),
	resync func(),
) {
	subscribed := false
	onSubscribed := func() {
		if subscribed {
			resync()
		}
		subscribed = true
	}
	for ctx.Err() == nil {
		if err := subscribeOnce(ctx, rdb, deliver, onSubscribed); err != nil {
			log.Warn("event_bus_subscribe_failed", zap.Error(err))
			select {
			case <-ctx.Done():
			case <-time.After(resubscribeDelay):
			}
		}
	}
}

func subscribeOnce(
	ctx context.Context,
	rdb *goredis.Client,
	deliver func(channel string, payload []byte),
	onSubscribed func(),
) error {
	sub := rdb.PSubscribe(ctx, channelPrefix+"*")
	// A blocked read ends only when the subscription closes: go-redis heeds a
	// deadline, not a cancellation.
	stop := context.AfterFunc(ctx, func() { _ = sub.Close() })
	defer func() {
		stop()
		_ = sub.Close()
	}()
	for {
		msg, err := sub.Receive(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return fmt.Errorf("receive events: %w", err)
		}
		switch m := msg.(type) {
		case *goredis.Subscription:
			onSubscribed()
		case *goredis.Message:
			deliver(m.Channel, []byte(m.Payload))
		}
	}
}
