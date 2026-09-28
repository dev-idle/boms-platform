// Package eventbus carries committed events over Redis Pub/Sub to every API
// process, one channel per user and per role. The channel names and the message
// shape here are the contract the realtime subscribers read.
package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
)

const channelPrefix = "boms:events:"

// userChannel carries events meant for one user.
func userChannel(id uuid.UUID) string {
	return channelPrefix + "user:" + id.String()
}

// roleChannel carries events meant for everyone holding a role.
func roleChannel(role domainuser.Role) string {
	return channelPrefix + "role:" + string(role)
}

// message is what a subscriber receives: which change happened and where to
// look again. It never carries the changed record.
type message struct {
	ID   string            `json:"id"`
	Type string            `json:"type"`
	At   time.Time         `json:"at"`
	Data map[string]string `json:"data"`
}

// RedisPublisher publishes events to their audiences' channels.
type RedisPublisher struct {
	rdb *goredis.Client
}

// NewRedisPublisher returns a publisher on rdb.
func NewRedisPublisher(rdb *goredis.Client) *RedisPublisher {
	return &RedisPublisher{rdb: rdb}
}

// Publish sends every event to each of its audience channels in one round trip.
func (p *RedisPublisher) Publish(ctx context.Context, events []domainevent.Event) error {
	if len(events) == 0 {
		return nil
	}
	pipe := p.rdb.Pipeline()
	for _, e := range events {
		payload, err := json.Marshal(message{
			ID:   e.ID.String(),
			Type: string(e.Topic),
			At:   e.OccurredAt,
			Data: e.Data,
		})
		if err != nil {
			return fmt.Errorf("encode event %s: %w", e.ID, err)
		}
		for _, id := range e.Audience.UserIDs {
			pipe.Publish(ctx, userChannel(id), payload)
		}
		for _, role := range e.Audience.Roles {
			pipe.Publish(ctx, roleChannel(role), payload)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("publish events: %w", err)
	}
	return nil
}

var _ port.EventPublisher = (*RedisPublisher)(nil)
