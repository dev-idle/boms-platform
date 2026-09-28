package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainstore "github.com/boms/backend/internal/domain/store"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// The migration seeds the rules that were constants before, and the row round
// trips what an admin saves.
func TestStoreSettingsRepository_Settings_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := newIntegrationPool(t, 2)
	repo := postgresadapter.NewStoreSettingsRepository(pool)

	seeded, err := repo.GetSettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, 8*time.Hour, seeded.OpensAt)
	assert.Equal(t, 18*time.Hour, seeded.ClosesAt)
	assert.Equal(t, 2*time.Hour, seeded.PreorderMinLead)
	assert.Equal(t, 14, seeded.MaxAdvanceDays)

	edited := domainstore.Settings{
		OpensAt: 7*time.Hour + 30*time.Minute, ClosesAt: 22 * time.Hour,
		PreorderMinLead: 90 * time.Minute, MaxAdvanceDays: 30,
	}
	var saved domainstore.Settings
	require.NoError(t, pool.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := repo.GetSettingsForUpdate(txCtx); err != nil {
			return err
		}
		saved, err = repo.UpdateSettings(txCtx, edited)
		return err
	}))
	assert.Equal(t, edited.OpensAt, saved.OpensAt)
	assert.Equal(t, edited.PreorderMinLead, saved.PreorderMinLead)
	assert.False(t, saved.UpdatedAt.IsZero())

	reread, err := repo.GetSettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, saved.ClosesAt, reread.ClosesAt)
	assert.Equal(t, saved.MaxAdvanceDays, reread.MaxAdvanceDays)

	_, err = repo.GetSettingsForUpdate(ctx)
	require.Error(t, err, "locking outside a transaction is refused")
}

// A day closes once; reopening it hides it from every list and frees the day.
func TestStoreSettingsRepository_ClosedDates_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := newIntegrationPool(t, 2)
	repo := postgresadapter.NewStoreSettingsRepository(pool)
	day := time.Date(2031, 2, 10, 0, 0, 0, 0, time.UTC)

	added, err := repo.AddClosedDate(ctx, day, "Lunar New Year")
	require.NoError(t, err)
	assert.Equal(t, day, added.Day)

	_, err = repo.AddClosedDate(ctx, day, "Again")
	require.ErrorIs(t, err, apperrors.ErrConflict)

	listed, err := repo.ListClosedDates(ctx, day.AddDate(0, 0, -1), day)
	require.NoError(t, err)
	require.Len(t, listed, 1)
	assert.Equal(t, "Lunar New Year", listed[0].Reason)

	outside, err := repo.ListClosedDates(ctx, day.AddDate(0, 0, 1), day.AddDate(0, 0, 30))
	require.NoError(t, err)
	assert.Empty(t, outside)

	removed, err := repo.RemoveClosedDate(ctx, added.ID)
	require.NoError(t, err)
	assert.Equal(t, added.ID, removed.ID)
	_, err = repo.RemoveClosedDate(ctx, added.ID)
	require.ErrorIs(t, err, apperrors.ErrNotFound)
	_, err = repo.RemoveClosedDate(ctx, uuid.New())
	require.ErrorIs(t, err, apperrors.ErrNotFound)

	listed, err = repo.ListClosedDates(ctx, day, day)
	require.NoError(t, err)
	assert.Empty(t, listed)
	_, err = repo.AddClosedDate(ctx, day, "Closed again")
	require.NoError(t, err, "a reopened day can close again")
}
