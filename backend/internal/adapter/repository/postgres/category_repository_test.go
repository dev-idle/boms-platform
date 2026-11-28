package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaincategory "github.com/boms/backend/internal/domain/category"
	"github.com/boms/backend/internal/port"
)

// The shop menu shows each category with how many products it holds and a
// cover photo. Only products the shop shows count, and the cover is the first
// photo of the first product by name that has one.
func TestCategoryRepository_CatalogListCountsAndCover_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	categories := postgresadapter.NewCategoryRepository(pool)
	products := postgresadapter.NewProductRepository(pool)

	newCategory := func(t *testing.T, slug string) uuid.UUID {
		t.Helper()
		category, err := categories.Create(ctx, port.CreateCategoryParams{
			Name: slug, Slug: slug, SortOrder: 10, IsActive: true, Station: domaincategory.StationKitchen,
		})
		require.NoError(t, err)
		return category.ID
	}
	newProduct := func(t *testing.T, categoryID uuid.UUID, name, slug string, active bool) uuid.UUID {
		t.Helper()
		product, err := products.Create(ctx, port.CreateProductParams{
			CategoryID: categoryID, Name: name, Slug: slug, PriceCents: 500, IsActive: active,
		})
		require.NoError(t, err)
		return product.ID
	}

	cakes := newCategory(t, "cover-cakes")
	empty := newCategory(t, "cover-empty")

	newProduct(t, cakes, "Apple tart", "cover-apple-tart", true)
	berry := newProduct(t, cakes, "Berry tart", "cover-berry-tart", true)
	cherry := newProduct(t, cakes, "Cherry tart", "cover-cherry-tart", true)
	hidden := newProduct(t, cakes, "Almond tart", "cover-almond-tart", false)

	berryFirst := "https://res.cloudinary.com/demo/image/upload/boms/products/berry-1.jpg"
	berrySecond := "https://res.cloudinary.com/demo/image/upload/boms/products/berry-2.jpg"
	require.NoError(t, products.ReplaceProductImages(ctx, berry, []string{berryFirst, berrySecond}))
	require.NoError(t, products.ReplaceProductImages(ctx, cherry, []string{"https://res.cloudinary.com/demo/image/upload/boms/products/cherry.jpg"}))
	require.NoError(t, products.ReplaceProductImages(ctx, hidden, []string{"https://res.cloudinary.com/demo/image/upload/boms/products/almond.jpg"}))

	listed, err := categories.CatalogList(ctx, port.CatalogListCategoriesParams{Limit: 100})
	require.NoError(t, err)
	byID := make(map[uuid.UUID]port.CatalogCategory, len(listed))
	for _, category := range listed {
		byID[category.ID] = category
	}

	t.Run("counts_only_shown_products_and_covers_with_the_first_photographed_one", func(t *testing.T) {
		got, ok := byID[cakes]
		require.True(t, ok)
		assert.EqualValues(t, 3, got.ProductCount)
		require.NotNil(t, got.ImageURL)
		assert.Equal(t, berryFirst, *got.ImageURL)
	})

	t.Run("an_empty_category_has_no_cover", func(t *testing.T) {
		got, ok := byID[empty]
		require.True(t, ok)
		assert.Zero(t, got.ProductCount)
		assert.Nil(t, got.ImageURL)
	})
}
