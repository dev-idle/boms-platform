package eventdispatch

import (
	"context"

	domainevent "github.com/boms/backend/internal/domain/event"
	"github.com/boms/backend/internal/port"
)

// Fanout delivers events to each publisher in turn and stops at the first that
// fails. The dispatcher then leaves the events pending, and the next delivery
// hands them to every publisher again — so each must take the same event twice:
// the realtime bus passes on a hint to refetch, and the email queue keeps one
// task per event.
func Fanout(publishers ...port.EventPublisher) port.EventPublisher {
	return fanout(publishers)
}

type fanout []port.EventPublisher

func (f fanout) Publish(ctx context.Context, events []domainevent.Event) error {
	for _, publisher := range f {
		if err := publisher.Publish(ctx, events); err != nil {
			return err
		}
	}
	return nil
}
