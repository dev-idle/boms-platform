package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

func TestOrderHistoryFilter(t *testing.T) {
	t.Parallel()

	t.Run("no_filter_is_the_whole_history", func(t *testing.T) {
		t.Parallel()
		filter, err := orderHistoryFilter(dto.OrderHistoryQuery{Status: " ", From: "", To: ""})
		require.NoError(t, err)
		assert.Nil(t, filter.Status)
		assert.Nil(t, filter.PlacedFrom)
		assert.Nil(t, filter.PlacedBefore)
	})

	t.Run("days_are_whole_bakery_days", func(t *testing.T) {
		t.Parallel()
		filter, err := orderHistoryFilter(dto.OrderHistoryQuery{Status: "ready", From: "2026-09-01", To: "2026-09-28"})
		require.NoError(t, err)
		require.NotNil(t, filter.Status)
		assert.Equal(t, domainorder.StatusReady, *filter.Status)
		require.NotNil(t, filter.PlacedFrom)
		require.NotNil(t, filter.PlacedBefore)
		assert.True(t, time.Date(2026, 9, 1, 0, 0, 0, 0, domainstore.Location).Equal(*filter.PlacedFrom))
		assert.True(t, time.Date(2026, 9, 29, 0, 0, 0, 0, domainstore.Location).Equal(*filter.PlacedBefore),
			"the last day is included up to its bakery midnight")
		assert.True(t, time.Date(2026, 9, 28, 17, 0, 0, 0, time.UTC).Equal(*filter.PlacedBefore),
			"bakery midnight is 17:00 UTC the day before")
	})

	t.Run("one_day", func(t *testing.T) {
		t.Parallel()
		filter, err := orderHistoryFilter(dto.OrderHistoryQuery{From: "2026-09-28", To: "2026-09-28"})
		require.NoError(t, err)
		assert.Equal(t, 24*time.Hour, filter.PlacedBefore.Sub(*filter.PlacedFrom))
	})

	for name, tc := range map[string]struct {
		query dto.OrderHistoryQuery
		field string
	}{
		"unknown_status":    {dto.OrderHistoryQuery{Status: "shipped"}, "status"},
		"unreadable_from":   {dto.OrderHistoryQuery{From: "28/09/2026"}, "from"},
		"unreadable_to":     {dto.OrderHistoryQuery{To: "2026-02-30"}, "to"},
		"to_before_from":    {dto.OrderHistoryQuery{From: "2026-09-28", To: "2026-09-27"}, "to"},
		"timestamp_as_from": {dto.OrderHistoryQuery{From: "2026-09-28T00:00:00Z"}, "from"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := orderHistoryFilter(tc.query)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, apperrors.ErrValidation.Code, appErr.Code)
			assert.Contains(t, appErr.Details, tc.field)
		})
	}
}
