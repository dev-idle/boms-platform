package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type fakeResource struct {
	name string
	err  error
}

func (f fakeResource) Name() string                   { return f.name }
func (f fakeResource) Ping(ctx context.Context) error { return f.err }

func TestReadiness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		resources  []port.HealthResource
		wantStatus string
		wantDB     string
		wantRedis  string
	}{
		{
			name: "all healthy",
			resources: []port.HealthResource{
				fakeResource{name: "postgres"},
				fakeResource{name: "redis"},
			},
			wantStatus: "ok",
			wantDB:     "up",
			wantRedis:  "up",
		},
		{
			name: "db down",
			resources: []port.HealthResource{
				fakeResource{name: "postgres", err: errors.New("connection refused")},
				fakeResource{name: "redis"},
			},
			wantStatus: "degraded",
			wantDB:     "down",
			wantRedis:  "up",
		},
		{
			name: "redis down",
			resources: []port.HealthResource{
				fakeResource{name: "postgres"},
				fakeResource{name: "redis", err: errors.New("timeout")},
			},
			wantStatus: "degraded",
			wantDB:     "up",
			wantRedis:  "down",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			log := zaptest.NewLogger(t)
			r := usecase.NewReadiness(tt.resources, time.Second, log)
			out := r.Execute(context.Background())
			assert.Equal(t, tt.wantStatus, out.Status)
			assert.Equal(t, tt.wantDB, out.Checks.DB)
			assert.Equal(t, tt.wantRedis, out.Checks.Redis)
			require.NotEmpty(t, out.Checks.DB)
		})
	}
}
