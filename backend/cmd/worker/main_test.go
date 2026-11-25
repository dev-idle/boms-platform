package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/config"
)

type fakeJobs struct {
	mu        sync.Mutex
	batches   []int
	sweeps    int
	prunes    int
	purges    int
	err       error
	retention time.Duration
}

func (f *fakeJobs) Sweep(context.Context, time.Duration, int32) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sweeps++
	if f.err != nil {
		return 0, f.err
	}
	if len(f.batches) == 0 {
		return 0, nil
	}
	next := f.batches[0]
	f.batches = f.batches[1:]
	return next, nil
}

func (f *fakeJobs) Prune(context.Context, time.Duration) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prunes++
	return 0, nil
}

func (f *fakeJobs) DeleteIdle(_ context.Context, retention time.Duration) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.purges++
	f.retention = retention
	return 0, nil
}

func (f *fakeJobs) counts() (sweeps, prunes int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sweeps, f.prunes
}

func (f *fakeJobs) purged() (int, time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.purges, f.retention
}

func TestSweep_DrainsABacklogInOneTick(t *testing.T) {
	t.Parallel()
	jobs := &fakeJobs{batches: []int{2, 2, 1}}

	sweep(context.Background(), jobs, config.OutboxConfig{SweepBatch: 2}, zap.NewNop())

	sweeps, _ := jobs.counts()
	assert.Equal(t, 3, sweeps, "full batches are followed at once; a short one ends the tick")
}

func TestSweep_StopsOnError(t *testing.T) {
	t.Parallel()
	jobs := &fakeJobs{err: errors.New("database down")}

	sweep(context.Background(), jobs, config.OutboxConfig{SweepBatch: 2}, zap.NewNop())

	sweeps, _ := jobs.counts()
	assert.Equal(t, 1, sweeps)
}

// beats counts the worker's recorded runs.
type beats struct{ count atomic.Int32 }

func (b *beats) Beat(context.Context, time.Time) error {
	b.count.Add(1)
	return nil
}

func (b *beats) LastBeat(context.Context) (*time.Time, error) { return nil, nil }

func TestRun_AStuckJobHoldsNoOtherBack(t *testing.T) {
	t.Parallel()
	var stuckEnded, others atomic.Int32
	orders := []orderJob{
		{name: "stuck", run: func(ctx context.Context) (int, error) {
			<-ctx.Done()
			stuckEnded.Add(1)
			return 0, ctx.Err()
		}},
		{name: "count", run: func(context.Context) (int, error) {
			others.Add(1)
			return 0, nil
		}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		run(ctx, &fakeJobs{}, &fakeJobs{}, orders, &beats{}, &config.Config{
			Outbox: config.OutboxConfig{SweepInterval: time.Hour, PruneInterval: time.Hour, SweepBatch: 10},
			Order:  config.OrderConfig{JobInterval: 10 * time.Millisecond},
		}, zap.NewNop())
		close(done)
	}()

	assert.Eventually(t, func() bool { return others.Load() >= 3 && stuckEnded.Load() >= 1 }, 5*time.Second, 5*time.Millisecond,
		"the other job keeps its rhythm, and the stuck run ends at its deadline")

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("run did not stop when its context was cancelled")
	}
}

func TestRun_SweepsPrunesPurgesAndExpiresUntilStopped(t *testing.T) {
	t.Parallel()
	jobs := &fakeJobs{}
	var orderRuns atomic.Int32
	orders := []orderJob{{name: "count", run: func(context.Context) (int, error) {
		orderRuns.Add(1)
		return 0, nil
	}}}
	heartbeat := &beats{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		run(ctx, jobs, jobs, orders, heartbeat, &config.Config{
			Outbox: config.OutboxConfig{
				SweepInterval: 5 * time.Millisecond,
				PruneInterval: 5 * time.Millisecond,
				SweepBatch:    10,
			},
			Order:     config.OrderConfig{JobInterval: 5 * time.Millisecond},
			Assistant: config.AssistantConfig{Retention: 60 * 24 * time.Hour},
		}, zap.NewNop())
		close(done)
	}()

	assert.Eventually(t, func() bool {
		sweeps, prunes := jobs.counts()
		purges, _ := jobs.purged()
		return sweeps > 0 && prunes > 0 && purges > 0 && orderRuns.Load() > 0 && heartbeat.count.Load() > 0
	}, 5*time.Second, 5*time.Millisecond)
	_, retention := jobs.purged()
	assert.Equal(t, 60*24*time.Hour, retention, "the assistant's conversations are kept as long as the policy says")

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("run did not stop when its context was cancelled")
	}
}
