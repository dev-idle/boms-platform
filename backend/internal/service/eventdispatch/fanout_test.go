package eventdispatch

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
)

func TestFanout(t *testing.T) {
	t.Parallel()
	customer := uuid.New()
	events := []domainevent.Event{domainorder.CreatedEvent(domainorder.Order{ID: uuid.New(), UserID: &customer})}

	t.Run("every_publisher_gets_the_events", func(t *testing.T) {
		t.Parallel()
		bus, emails := &fakePublisher{}, &fakePublisher{}

		require.NoError(t, Fanout(bus, emails).Publish(context.Background(), events))

		assert.Equal(t, events, bus.sent)
		assert.Equal(t, events, emails.sent)
	})

	t.Run("stops_at_the_first_failure", func(t *testing.T) {
		t.Parallel()
		down := errors.New("bus down")
		bus, emails := &fakePublisher{err: down}, &fakePublisher{}

		err := Fanout(bus, emails).Publish(context.Background(), events)

		require.ErrorIs(t, err, down)
		assert.Empty(t, emails.sent, "the events stay pending and reach every publisher next time")
	})
}
