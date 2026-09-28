package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domaincombo "github.com/boms/backend/internal/domain/combo"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// Combo reads take uuid[] parameters and carry a nullable image column — the
// shapes the driver has to encode and decode itself, so they are checked
// against a real database rather than read off the generated code.
func TestComboRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	categories := postgresadapter.NewCategoryRepository(pool)
	products := postgresadapter.NewProductRepository(pool)
	combos := postgresadapter.NewComboRepository(pool)

	category, err := categories.Create(ctx, port.CreateCategoryParams{
		Name: "Boxes", Slug: "boxes", SortOrder: 1, IsActive: true, Station: domaincategory.StationCounter,
	})
	require.NoError(t, err)
	croissant, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: category.ID, Name: "Croissant", Slug: "croissant", PriceCents: 300, IsActive: true,
	})
	require.NoError(t, err)
	eclair, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: category.ID, Name: "Eclair", Slug: "eclair", PriceCents: 450, IsActive: true,
	})
	require.NoError(t, err)

	// newBox creates a live two-product combo of its own for one subtest.
	newBox := func(t *testing.T, slug string, imageURL *string) *domaincombo.Combo {
		t.Helper()
		now := time.Now().UTC()
		combo, err := combos.Create(ctx, port.CreateComboParams{
			Name:       "Box " + slug,
			Slug:       slug,
			PriceCents: 650,
			ImageURL:   imageURL,
			StartsAt:   now.Add(-time.Hour),
			EndsAt:     now.Add(24 * time.Hour),
			IsActive:   true,
		})
		require.NoError(t, err)
		require.NoError(t, combos.ReplaceItems(ctx, combo.ID, []port.ComboItemParams{
			{ProductID: croissant.ID, Quantity: 2},
			{ProductID: eclair.ID, Quantity: 1},
		}))
		return combo
	}

	t.Run("keeps_a_missing_image_and_deletion_as_nil", func(t *testing.T) {
		combo := newBox(t, "no-image-box", nil)
		got, err := combos.GetByID(ctx, combo.ID)
		require.NoError(t, err)
		assert.Nil(t, got.ImageURL)
		assert.Nil(t, got.DeletedAt)
	})

	t.Run("round_trips_an_image_url", func(t *testing.T) {
		image := "https://res.cloudinary.com/demo/image/upload/boms/combos/box.jpg"
		combo := newBox(t, "image-box", &image)
		got, err := combos.GetByID(ctx, combo.ID)
		require.NoError(t, err)
		require.NotNil(t, got.ImageURL)
		assert.Equal(t, image, *got.ImageURL)
	})

	t.Run("reads_items_for_a_list_of_combo_ids", func(t *testing.T) {
		combo := newBox(t, "items-box", nil)
		items, err := combos.ListItemsByComboIDs(ctx, []uuid.UUID{combo.ID, uuid.New()})
		require.NoError(t, err)
		require.Len(t, items[combo.ID], 2)
	})

	t.Run("counts_available_products_from_a_list_of_ids", func(t *testing.T) {
		count, err := combos.CountAvailableProducts(ctx, []uuid.UUID{croissant.ID, eclair.ID, uuid.New()})
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("serves_an_active_combo_to_the_catalog_by_id", func(t *testing.T) {
		combo := newBox(t, "catalog-box", nil)
		got, err := combos.CatalogGetByIDs(ctx, []uuid.UUID{combo.ID})
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Nil(t, got[0].ImageURL)
		assert.Len(t, got[0].Items, 2)
	})

	t.Run("hides_a_soft_deleted_combo", func(t *testing.T) {
		combo := newBox(t, "deleted-box", nil)
		require.NoError(t, combos.SoftDelete(ctx, combo.ID))
		_, err := combos.GetByID(ctx, combo.ID)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}
