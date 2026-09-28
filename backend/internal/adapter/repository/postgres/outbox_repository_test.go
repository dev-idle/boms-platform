package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
)

func newOrderEvent() domainevent.Event {
	return domainorder.StatusChangedEvent(domainorder.StatusPending,
		domainorder.Order{ID: uuid.New(), UserID: uuid.New(), Status: domainorder.StatusConfirmed})
}

func eventIDs(events []domainevent.Event) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.ID)
	}
	return ids
}

// record commits events in one transaction.
func record(t *testing.T, pool *postgresadapter.Pool, outbox *postgresadapter.OutboxRepository, events ...domainevent.Event) {
	t.Helper()
	require.NoError(t, pool.WithTx(context.Background(), func(txCtx context.Context) error {
		for _, e := range events {
			if err := outbox.Add(txCtx, e); err != nil {
				return err
			}
		}
		return nil
	}))
}

// claim returns the pending events at least grace old, in a transaction of its own.
func claim(t *testing.T, pool *postgresadapter.Pool, outbox *postgresadapter.OutboxRepository, grace time.Duration) []domainevent.Event {
	t.Helper()
	var got []domainevent.Event
	require.NoError(t, pool.WithTx(context.Background(), func(txCtx context.Context) error {
		var err error
		got, err = outbox.ClaimUnpublished(txCtx, grace, 100)
		return err
	}))
	return got
}

// The outbox guarantees an event exists exactly when its change committed, and
// that concurrent sweepers never hold the same pending event at once. Each
// subtest marks what it recorded as published before it ends, so pending rows
// never leak into the next one.
func TestOutboxRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	outbox := postgresadapter.NewOutboxRepository(pool)

	var mu sync.Mutex
	var committed []domainevent.Event
	pool.OnCommit(func(_ context.Context, events []domainevent.Event) {
		mu.Lock()
		defer mu.Unlock()
		committed = append(committed, events...)
	})
	committedEvents := func() []domainevent.Event {
		mu.Lock()
		defer mu.Unlock()
		return append([]domainevent.Event(nil), committed...)
	}

	t.Run("refuses_an_event_outside_a_transaction", func(t *testing.T) {
		assert.Error(t, outbox.Add(ctx, newOrderEvent()))
	})

	t.Run("hands_committed_events_to_the_hook_stamped_by_the_database", func(t *testing.T) {
		e := newOrderEvent()
		record(t, pool, outbox, e)
		t.Cleanup(func() { _ = outbox.MarkPublished(ctx, []uuid.UUID{e.ID}) })

		var handed *domainevent.Event
		for _, got := range committedEvents() {
			if got.ID == e.ID {
				handed = &got
			}
		}
		require.NotNil(t, handed, "the commit hook receives the event")
		assert.WithinDuration(t, time.Now(), handed.OccurredAt, time.Minute, "OccurredAt comes from the database clock")

		stored := claim(t, pool, outbox, 0)
		require.Contains(t, eventIDs(stored), e.ID)
		for _, got := range stored {
			if got.ID == e.ID {
				assert.Equal(t, e.Topic, got.Topic)
				assert.Equal(t, e.Data, got.Data)
				assert.Equal(t, e.Audience.UserIDs, got.Audience.UserIDs)
				assert.ElementsMatch(t, []domainuser.Role{domainuser.RoleStaff, domainuser.RoleBaker}, got.Audience.Roles)
				assert.True(t, handed.OccurredAt.Equal(got.OccurredAt), "hook and store agree on when it happened")
			}
		}
	})

	t.Run("keeps_a_public_audience", func(t *testing.T) {
		e := domainstore.SettingsUpdatedEvent()
		record(t, pool, outbox, e)
		t.Cleanup(func() { _ = outbox.MarkPublished(ctx, []uuid.UUID{e.ID}) })

		for _, got := range claim(t, pool, outbox, 0) {
			if got.ID == e.ID {
				assert.True(t, got.Audience.Public)
				assert.Empty(t, got.Audience.UserIDs)
				return
			}
		}
		t.Fatal("the public event was not stored")
	})

	t.Run("announces_nothing_for_a_change_that_rolled_back", func(t *testing.T) {
		e := newOrderEvent()
		err := pool.WithTx(ctx, func(txCtx context.Context) error {
			if err := outbox.Add(txCtx, e); err != nil {
				return err
			}
			return errors.New("change failed")
		})
		require.Error(t, err)

		assert.NotContains(t, eventIDs(committedEvents()), e.ID)
		assert.NotContains(t, eventIDs(claim(t, pool, outbox, 0)), e.ID)
	})

	t.Run("keeps_a_failed_event_pending_and_stops_offering_a_published_one", func(t *testing.T) {
		failed, delivered := newOrderEvent(), newOrderEvent()
		record(t, pool, outbox, failed, delivered)
		t.Cleanup(func() { _ = outbox.MarkPublished(ctx, []uuid.UUID{failed.ID}) })

		require.NoError(t, outbox.RecordFailure(ctx, []uuid.UUID{failed.ID}, "bus down"))
		require.NoError(t, outbox.MarkPublished(ctx, []uuid.UUID{delivered.ID}))

		pending := eventIDs(claim(t, pool, outbox, 0))
		assert.Contains(t, pending, failed.ID)
		assert.NotContains(t, pending, delivered.ID)
	})

	t.Run("leaves_events_younger_than_the_grace_to_the_post_commit_delivery", func(t *testing.T) {
		fresh := newOrderEvent()
		record(t, pool, outbox, fresh)
		t.Cleanup(func() { _ = outbox.MarkPublished(ctx, []uuid.UUID{fresh.ID}) })

		assert.NotContains(t, eventIDs(claim(t, pool, outbox, time.Minute)), fresh.ID)
		assert.Contains(t, eventIDs(claim(t, pool, outbox, 0)), fresh.ID)
	})

	t.Run("concurrent_sweeps_never_claim_the_same_event", func(t *testing.T) {
		e := newOrderEvent()
		record(t, pool, outbox, e)
		t.Cleanup(func() { _ = outbox.MarkPublished(ctx, []uuid.UUID{e.ID}) })

		held := make(chan struct{})
		release := make(chan struct{})
		firstDone := make(chan error, 1)
		go func() {
			firstDone <- pool.WithTx(ctx, func(txCtx context.Context) error {
				claimed, err := outbox.ClaimUnpublished(txCtx, 0, 100)
				if err != nil {
					return err
				}
				if len(claimed) == 0 {
					return errors.New("first sweeper claimed nothing")
				}
				close(held)
				<-release
				return nil
			})
		}()

		<-held
		assert.NotContains(t, eventIDs(claim(t, pool, outbox, 0)), e.ID, "a row one sweeper holds is skipped by the next")
		close(release)
		require.NoError(t, <-firstDone)
	})
}

// Retention is measured from publication: a row is deleted only once it has
// been published for longer than the retention, and a pending row never is.
func TestOutboxRepository_DeletePublished_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 2)
	outbox := postgresadapter.NewOutboxRepository(pool)

	published, pending := newOrderEvent(), newOrderEvent()
	record(t, pool, outbox, published, pending)
	require.NoError(t, outbox.MarkPublished(ctx, []uuid.UUID{published.ID}))

	deleted, err := outbox.DeletePublished(ctx, time.Hour)
	require.NoError(t, err)
	assert.Zero(t, deleted, "published moments ago, inside the retention")

	time.Sleep(20 * time.Millisecond)
	deleted, err = outbox.DeletePublished(ctx, 10*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted, "past the retention")

	assert.Equal(t, []uuid.UUID{pending.ID}, eventIDs(claim(t, pool, outbox, 0)), "an undelivered event is never pruned")
}
