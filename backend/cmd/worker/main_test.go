package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/boms/backend/internal/config"
)

type fakeJobs struct {
	mu      sync.Mutex
	batches []int
	sweeps  int
	prunes  int
	err     error
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

func (f *fakeJobs) counts() (sweeps, prunes int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sweeps, f.prunes
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

func TestRun_SweepsAndPrunesUntilStopped(t *testing.T) {
	t.Parallel()
	jobs := &fakeJobs{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		run(ctx, jobs, config.OutboxConfig{
			SweepInterval: 5 * time.Millisecond,
			PruneInterval: 5 * time.Millisecond,
			SweepBatch:    10,
		}, zap.NewNop())
		close(done)
	}()

	assert.Eventually(t, func() bool {
		sweeps, prunes := jobs.counts()
		return sweeps > 0 && prunes > 0
	}, 5*time.Second, 5*time.Millisecond)

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("run did not stop when its context was cancelled")
	}
}
