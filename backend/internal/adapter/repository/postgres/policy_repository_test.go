package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
)

// A policy version keeps the time an API first served it, however often the
// API starts again: the overlap for the version before it never starts over.
func TestPolicyRepository_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	policies := postgresadapter.NewPolicyRepository(newIntegrationPool(t, 2))
	version := "test-" + uuid.NewString()[:8]

	first, err := policies.Publish(ctx, version)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), first, time.Minute)

	again, err := policies.Publish(ctx, version)
	require.NoError(t, err)
	assert.True(t, first.Equal(again), "a restart keeps the first time")

	next, err := policies.Publish(ctx, version+"-2")
	require.NoError(t, err)
	assert.False(t, next.Before(first), "a new version starts its own time")
}
