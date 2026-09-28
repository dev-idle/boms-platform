package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/usecase"
)

func TestStoreUsecase_PickupRules(t *testing.T) {
	t.Parallel()
	store := newMemoryStore()
	today := domainstore.DayOf(time.Now())
	_, _ = store.AddClosedDate(context.Background(), today.AddDate(0, 0, 3), "Inventory")
	_, _ = store.AddClosedDate(context.Background(), today.AddDate(0, 0, 40), "Beyond the booking window")

	out, err := usecase.NewStoreUsecase(store).PickupRules(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "08:00", out.OpensAt)
	assert.Equal(t, "18:00", out.ClosesAt)
	assert.Equal(t, 120, out.PreorderMinLeadMinutes)
	assert.Equal(t, 14, out.MaxAdvanceDays)
	assert.Equal(t, []dto.PublicClosedDateResponse{
		{Date: today.AddDate(0, 0, 3).Format(domainstore.DayLayout), Reason: "Inventory"},
	}, out.ClosedDates, "only closures a customer could pick are listed")
}

func TestStoreUsecase_PickupRules_StoreDown(t *testing.T) {
	t.Parallel()
	store := newMemoryStore()
	store.err = errors.New("database unavailable")

	_, err := usecase.NewStoreUsecase(store).PickupRules(context.Background())

	require.ErrorIs(t, err, store.err)
}
