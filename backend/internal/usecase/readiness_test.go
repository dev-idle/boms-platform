package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

type fakeResource struct {
	name string
	err  error
}

func (f fakeResource) Name() string                   { return f.name }
func (f fakeResource) Ping(ctx context.Context) error { return f.err }

func TestReadiness_AllHealthy(t *testing.T) {
	t.Parallel()
	r := usecase.NewReadiness([]port.HealthResource{
		fakeResource{name: "a"},
		fakeResource{name: "b"},
	}, time.Second)
	out := r.Execute(context.Background())
	if out.Status != "ready" || len(out.Checks) != 2 {
		t.Fatalf("unexpected: %+v", out)
	}
	for _, c := range out.Checks {
		if !c.OK {
			t.Fatalf("check %q not ok: %s", c.Name, c.Error)
		}
	}
}

func TestReadiness_OneDependencyFails(t *testing.T) {
	t.Parallel()
	r := usecase.NewReadiness([]port.HealthResource{
		fakeResource{name: "ok"},
		fakeResource{name: "bad", err: errors.New("down")},
	}, time.Second)
	out := r.Execute(context.Background())
	if out.Status != "not_ready" {
		t.Fatalf("status=%q", out.Status)
	}
	var sawBad bool
	for _, c := range out.Checks {
		if c.Name == "bad" {
			sawBad = true
			if c.OK || c.Error == "" {
				t.Fatalf("bad check: %+v", c)
			}
		}
	}
	if !sawBad {
		t.Fatal("missing bad check")
	}
}
