package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// countedEngagement answers with counts, or err.
type countedEngagement struct {
	counts port.EngagementCounts
	err    error
}

func (f countedEngagement) CountEngagement(context.Context) (port.EngagementCounts, error) {
	return f.counts, f.err
}

func TestManagerEngagementUsecase_Report(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("lists_each_feature_in_a_fixed_order", func(t *testing.T) {
		t.Parallel()
		uc := usecase.NewManagerEngagementUsecase(countedEngagement{counts: port.EngagementCounts{
			Customers: 10, Reviews: 1, Favorites: 2, Wishlist: 3, Messages: 4, Promotions: 5,
			UsedNone: 3, UsedOne: 3, UsedTwoOrMore: 4,
		}})

		out, err := uc.Report(ctx)

		require.NoError(t, err)
		assert.Equal(t, int64(10), out.Customers)
		assert.Equal(t, []int64{3, 3, 4}, []int64{out.UsedNone, out.UsedOne, out.UsedTwoOrMore})
		var features []string
		var customers []int64
		for _, feature := range out.Features {
			features = append(features, feature.Feature)
			customers = append(customers, feature.Customers)
		}
		assert.Equal(t, []string{"reviews", "favorites", "wishlist", "messages", "promotions"}, features)
		assert.Equal(t, []int64{1, 2, 3, 4, 5}, customers)
	})

	t.Run("a_failed_read_is_reported", func(t *testing.T) {
		t.Parallel()
		down := errors.New("database unavailable")

		_, err := usecase.NewManagerEngagementUsecase(countedEngagement{err: down}).Report(ctx)

		require.ErrorIs(t, err, down)
	})
}
