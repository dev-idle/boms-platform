// Package eventdispatch delivers outbox events. Right after a transaction
// commits, its events are published in the background; a worker sweeps up any
// that a crash or a bus outage left behind. Delivery is at least once and in no
// guaranteed order across transactions: an event can arrive twice, and
// subscribers only use it as a hint to refetch.
package eventdispatch

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainevent "github.com/boms/backend/internal/domain/event"
	"github.com/boms/backend/internal/port"
)

// Dispatcher publishes outbox events and records the outcome on their rows.
type Dispatcher struct {
	store     port.OutboxStore
	publisher port.EventPublisher
	tx        port.TxManager
	log       *zap.Logger
	timeout   time.Duration
	inflight  sync.WaitGroup
}

// New returns a dispatcher. timeout bounds each background publish, and
// separately the write that records how it went — so a background delivery is
// over within twice that.
func New(
	store port.OutboxStore,
	publisher port.EventPublisher,
	tx port.TxManager,
	log *zap.Logger,
	timeout time.Duration,
) *Dispatcher {
	return &Dispatcher{store: store, publisher: publisher, tx: tx, log: log, timeout: timeout}
}

// AfterCommit delivers events a transaction has just committed without holding
// up the request that made the change. The change is already durable, so a
// failed delivery is only recorded; the sweeper sends it again.
func (d *Dispatcher) AfterCommit(ctx context.Context, events []domainevent.Event) {
	d.inflight.Go(func() {
		publishCtx, cancel := d.detached(ctx)
		defer cancel()
		d.dispatch(ctx, publishCtx, events)
	})
}

// Wait blocks until background deliveries finish or ctx is done, so a shutting
// down process does not cut them off mid-flight.
func (d *Dispatcher) Wait(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		d.inflight.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		d.log.Warn("event_deliveries_left_to_sweeper", zap.Error(ctx.Err()))
	}
}

// dispatch publishes events on publishCtx and records the outcome on a fresh
// deadline: a publish that ran out of time must still be written down as failed.
func (d *Dispatcher) dispatch(parent, publishCtx context.Context, events []domainevent.Event) {
	ids := eventIDs(events)
	publishErr := d.publisher.Publish(publishCtx, events)

	recordCtx, cancel := d.detached(parent)
	defer cancel()
	if publishErr != nil {
		d.log.Warn("event_publish_failed", zap.Int("events", len(events)), zap.Error(publishErr))
		if err := d.store.RecordFailure(recordCtx, ids, publishErr.Error()); err != nil {
			d.log.Error("event_failure_record_failed", zap.Error(err))
		}
		return
	}
	if err := d.store.MarkPublished(recordCtx, ids); err != nil {
		// Delivered but still marked pending: the sweeper will send it again,
		// which subscribers tolerate.
		d.log.Warn("event_mark_published_failed", zap.Int("events", len(events)), zap.Error(err))
	}
}

// detached returns a context that outlives the request ctx belongs to but
// still ends after the dispatcher's timeout.
func (d *Dispatcher) detached(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), d.timeout)
}

// Sweep publishes up to batch events that are still unpublished grace after
// they were recorded — long enough that an AfterCommit delivery has finished or
// failed. It returns how many were delivered. A failure to mark them rolls the
// claim back, so they stay pending and are sent again.
func (d *Dispatcher) Sweep(ctx context.Context, grace time.Duration, batch int32) (int, error) {
	delivered := 0
	err := d.tx.WithTx(ctx, func(txCtx context.Context) error {
		events, err := d.store.ClaimUnpublished(txCtx, grace, batch)
		if err != nil || len(events) == 0 {
			return err
		}
		ids := eventIDs(events)
		if err := d.publisher.Publish(txCtx, events); err != nil {
			d.log.Warn("event_sweep_publish_failed", zap.Int("events", len(events)), zap.Error(err))
			return d.store.RecordFailure(txCtx, ids, err.Error())
		}
		if err := d.store.MarkPublished(txCtx, ids); err != nil {
			return err
		}
		delivered = len(events)
		return nil
	})
	return delivered, err
}

// Prune deletes delivery records published longer ago than retention.
func (d *Dispatcher) Prune(ctx context.Context, retention time.Duration) (int64, error) {
	return d.store.DeletePublished(ctx, retention)
}

func eventIDs(events []domainevent.Event) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(events))
	for _, e := range events {
		ids = append(ids, e.ID)
	}
	return ids
}
