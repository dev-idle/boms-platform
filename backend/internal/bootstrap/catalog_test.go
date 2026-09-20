package bootstrap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundlePriceCents(t *testing.T) {
	t.Parallel()

	t.Run("rounds_down_to_a_whole_dime", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, int64(2350), bundlePriceCents(2500, 6))
	})

	t.Run("skips_rounding_when_it_would_pass_the_house_ceiling", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, int64(5), bundlePriceCents(5, 8))
		assert.Equal(t, int64(96), bundlePriceCents(100, 4))
	})
}

func TestSeedCombosStayWithinTheHouseDiscountRule(t *testing.T) {
	t.Parallel()

	prices := make(map[string]int64, len(seedProducts))
	for _, product := range seedProducts {
		prices[product.Slug] = product.PriceCents
	}

	for _, combo := range seedCombos {
		t.Run(combo.Slug, func(t *testing.T) {
			t.Parallel()

			var sum int64
			for _, item := range combo.Items {
				price, ok := prices[item.ProductSlug]
				require.True(t, ok, "combo references unknown product %q", item.ProductSlug)
				sum += price * int64(item.Quantity)
			}

			saved := sum - bundlePriceCents(sum, combo.DiscountPercent)
			percent := float64(saved) / float64(sum) * 100
			assert.GreaterOrEqual(t, percent, 3.0)
			assert.LessOrEqual(t, percent, 8.0)
		})
	}
}

func TestWindowRange(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, time.September, 20, 9, 0, 0, 0, time.UTC)

	t.Run("live_window_covers_now", func(t *testing.T) {
		t.Parallel()
		startsAt, endsAt := windowRange(windowLive, now)
		assert.True(t, startsAt.Before(now) && endsAt.After(now))
	})

	t.Run("upcoming_window_opens_later", func(t *testing.T) {
		t.Parallel()
		startsAt, _ := windowRange(windowUpcoming, now)
		assert.True(t, startsAt.After(now))
	})

	t.Run("closed_window_already_ended", func(t *testing.T) {
		t.Parallel()
		_, endsAt := windowRange(windowClosed, now)
		assert.True(t, endsAt.Before(now))
	})
}

func TestSeedDataIsInternallyConsistent(t *testing.T) {
	t.Parallel()

	categories := make(map[string]struct{}, len(seedCategories))
	for _, category := range seedCategories {
		_, duplicate := categories[category.Slug]
		require.False(t, duplicate, "duplicate category slug %q", category.Slug)
		categories[category.Slug] = struct{}{}
	}

	slugs := make(map[string]struct{}, len(seedProducts))
	for _, product := range seedProducts {
		_, duplicate := slugs[product.Slug]
		require.False(t, duplicate, "duplicate product slug %q", product.Slug)
		slugs[product.Slug] = struct{}{}

		_, known := categories[product.CategorySlug]
		assert.True(t, known, "product %q points at unknown category %q", product.Slug, product.CategorySlug)
		assert.Positive(t, product.PriceCents, "product %q needs a price", product.Slug)
	}

	for _, combo := range seedCombos {
		seen := make(map[string]struct{}, len(combo.Items))
		for _, item := range combo.Items {
			_, duplicate := seen[item.ProductSlug]
			require.False(t, duplicate, "combo %q repeats product %q", combo.Slug, item.ProductSlug)
			seen[item.ProductSlug] = struct{}{}
			assert.Positive(t, item.Quantity)
		}
	}

	for _, code := range seedDiscountCodes {
		if code.MaxUses != nil {
			assert.LessOrEqual(t, code.UsedCount, *code.MaxUses, "code %q used more than its cap", code.Code)
		}
		if code.MaxDiscountCents != nil {
			assert.Equal(t, "percent", string(code.DiscountType), "code %q caps a non-percent discount", code.Code)
		}
	}
}
