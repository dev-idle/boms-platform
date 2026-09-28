package eventdispatch

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
)

type fakeStore struct {
	mu         sync.Mutex
	pending    []domainevent.Event
	claimGrace time.Duration
	published  []uuid.UUID
	failed     []uuid.UUID
	failReason string
	markErr    error
	recordErr  error
	retention  time.Duration
	// recordCtxErr is the state of the context RecordFailure was given.
	recordCtxErr error
}

func (f *fakeStore) ClaimUnpublished(_ context.Context, grace time.Duration, _ int32) ([]domainevent.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.claimGrace = grace
	return f.pending, nil
}

func (f *fakeStore) MarkPublished(_ context.Context, ids []uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.markErr != nil {
		return f.markErr
	}
	f.published = append(f.published, ids...)
	return nil
}

func (f *fakeStore) RecordFailure(ctx context.Context, ids []uuid.UUID, reason string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recordCtxErr = ctx.Err()
	if f.recordErr != nil {
		return f.recordErr
	}
	f.failed = append(f.failed, ids...)
	f.failReason = reason
	return nil
}

func (f *fakeStore) DeletePublished(_ context.Context, retention time.Duration) (int64, error) {
	f.retention = retention
	return 3, nil
}

func (f *fakeStore) snapshot() (published, failed []uuid.UUID) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]uuid.UUID(nil), f.published...), append([]uuid.UUID(nil), f.failed...)
}

type fakePublisher struct {
	mu   sync.Mutex
	err  error
	sent []domainevent.Event
	// hang makes Publish wait until its context is done, like a stalled Redis.
	hang bool
}

func (f *fakePublisher) Publish(ctx context.Context, events []domainevent.Event) error {
	if f.hang {
		<-ctx.Done()
		return ctx.Err()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return f.err
	}
	f.sent = append(f.sent, events...)
	return nil
}

// inlineTx runs fn directly; the fakes hold no state a rollback would undo.
type inlineTx struct{}

func (inlineTx) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

func newEvent() domainevent.Event {
	return domainorder.StatusChangedEvent(domainorder.StatusPending,
		domainorder.Order{ID: uuid.New(), UserID: uuid.New(), Status: domainorder.StatusConfirmed})
}

func newDispatcher(store *fakeStore, publisher *fakePublisher, timeout time.Duration) *Dispatcher {
	return New(store, publisher, inlineTx{}, zap.NewNop(), timeout)
}

// deliver runs one AfterCommit delivery to completion.
func deliver(t *testing.T, d *Dispatcher, events ...domainevent.Event) {
	t.Helper()
	d.AfterCommit(context.Background(), events)
	waitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	d.Wait(waitCtx)
	require.NoError(t, waitCtx.Err(), "background delivery never finished")
}

func TestDispatcher_AfterCommit(t *testing.T) {
	t.Parallel()

	t.Run("marks_delivered_events_published", func(t *testing.T) {
		t.Parallel()
		store, publisher := &fakeStore{}, &fakePublisher{}
		e := newEvent()

		deliver(t, newDispatcher(store, publisher, time.Second), e)

		published, failed := store.snapshot()
		assert.Equal(t, []uuid.UUID{e.ID}, published)
		assert.Empty(t, failed)
	})

	t.Run("records_a_failed_delivery_for_the_sweeper", func(t *testing.T) {
		t.Parallel()
		store := &fakeStore{}
		e := newEvent()

		deliver(t, newDispatcher(store, &fakePublisher{err: errors.New("bus down")}, time.Second), e)

		published, failed := store.snapshot()
		assert.Empty(t, published)
		assert.Equal(t, []uuid.UUID{e.ID}, failed)
		assert.Contains(t, store.failReason, "bus down")
	})

	t.Run("records_a_delivery_that_ran_out_of_time_on_a_fresh_deadline", func(t *testing.T) {
		t.Parallel()
		store := &fakeStore{}
		e := newEvent()

		deliver(t, newDispatcher(store, &fakePublisher{hang: true}, 20*time.Millisecond), e)

		_, failed := store.snapshot()
		assert.Equal(t, []uuid.UUID{e.ID}, failed)
		assert.NoError(t, store.recordCtxErr, "the failure must be written after the publish deadline passed")
	})

	t.Run("delivers_after_the_request_that_committed_has_ended", func(t *testing.T) {
		t.Parallel()
		store := &fakeStore{}
		d := newDispatcher(store, &fakePublisher{}, time.Second)
		e := newEvent()

		requestCtx, endRequest := context.WithCancel(context.Background())
		d.AfterCommit(requestCtx, []domainevent.Event{e})
		endRequest()
		waitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		d.Wait(waitCtx)

		published, _ := store.snapshot()
		assert.Equal(t, []uuid.UUID{e.ID}, published)
	})

	t.Run("survives_failing_to_mark_or_record", func(t *testing.T) {
		t.Parallel()
		markFails := &fakeStore{markErr: errors.New("db down")}
		deliver(t, newDispatcher(markFails, &fakePublisher{}, time.Second), newEvent())

		recordFails := &fakeStore{recordErr: errors.New("db down")}
		deliver(t, newDispatcher(recordFails, &fakePublisher{err: errors.New("bus down")}, time.Second), newEvent())
		// Both leave the row pending for the sweeper; neither may panic or block.
	})
}

func TestDispatcher_Wait_GivesUpAtItsDeadline(t *testing.T) {
	t.Parallel()
	d := newDispatcher(&fakeStore{}, &fakePublisher{hang: true}, time.Minute)
	d.AfterCommit(context.Background(), []domainevent.Event{newEvent()})

	waitCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	start := time.Now()
	d.Wait(waitCtx)
	assert.Less(t, time.Since(start), 5*time.Second, "shutdown must not wait on a stalled delivery")
}

func TestDispatcher_Sweep(t *testing.T) {
	t.Parallel()

	t.Run("delivers_what_is_still_pending_after_the_grace_period", func(t *testing.T) {
		t.Parallel()
		e := newEvent()
		store := &fakeStore{pending: []domainevent.Event{e}}

		delivered, err := newDispatcher(store, &fakePublisher{}, time.Second).Sweep(context.Background(), 15*time.Second, 100)
		require.NoError(t, err)
		assert.Equal(t, 1, delivered)
		assert.Equal(t, 15*time.Second, store.claimGrace)
		published, _ := store.snapshot()
		assert.Equal(t, []uuid.UUID{e.ID}, published)
	})

	t.Run("keeps_the_events_pending_when_the_bus_is_down", func(t *testing.T) {
		t.Parallel()
		e := newEvent()
		store := &fakeStore{pending: []domainevent.Event{e}}

		delivered, err := newDispatcher(store, &fakePublisher{err: errors.New("bus down")}, time.Second).
			Sweep(context.Background(), 15*time.Second, 100)
		require.NoError(t, err, "a bus outage is recorded, not a sweep failure")
		assert.Zero(t, delivered)
		published, failed := store.snapshot()
		assert.Empty(t, published)
		assert.Equal(t, []uuid.UUID{e.ID}, failed)
	})

	t.Run("fails_the_sweep_when_the_claim_cannot_be_marked", func(t *testing.T) {
		t.Parallel()
		errMark := errors.New("db down")
		store := &fakeStore{pending: []domainevent.Event{newEvent()}, markErr: errMark}

		delivered, err := newDispatcher(store, &fakePublisher{}, time.Second).Sweep(context.Background(), time.Second, 100)
		assert.ErrorIs(t, err, errMark, "the claim rolls back so the events stay pending")
		assert.Zero(t, delivered)
	})

	t.Run("does_nothing_when_nothing_is_pending", func(t *testing.T) {
		t.Parallel()
		publisher := &fakePublisher{}
		delivered, err := newDispatcher(&fakeStore{}, publisher, time.Second).Sweep(context.Background(), time.Second, 100)
		require.NoError(t, err)
		assert.Zero(t, delivered)
		assert.Empty(t, publisher.sent)
	})
}

func TestDispatcher_Prune(t *testing.T) {
	t.Parallel()
	store := &fakeStore{}

	deleted, err := newDispatcher(store, &fakePublisher{}, time.Second).Prune(context.Background(), 7*24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, int64(3), deleted)
	assert.Equal(t, 7*24*time.Hour, store.retention)
}
