package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The catalog and manager list queries carry the gallery with each row. These
// are the cases that shape cannot be checked by reading: the text[] column has
// to survive the driver, come back in sort order, and be empty — not null — for
// a product with no images.
func TestProductRepository_ListGallery_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	t.Parallel()

	ctx := context.Background()
	pgContainer, connStr, err := startPostgres(ctx, t)
	if err != nil {
		if strings.Contains(err.Error(), "docker") {
			t.Skip("docker not available")
		}
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })
	require.NoError(t, applyMigrations(ctx, connStr))

	pool, err := postgresadapter.NewPool(ctx, config.PostgresConfig{
		URL:                connStr,
		MaxConns:           5,
		MinConns:           1,
		MaxConnLifetime:    time.Hour,
		MaxConnIdleTime:    time.Minute,
		HealthCheckTimeout: 5 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	categories := postgresadapter.NewCategoryRepository(pool)
	products := postgresadapter.NewProductRepository(pool)

	category, err := categories.Create(ctx, port.CreateCategoryParams{
		Name:      "Gallery",
		Slug:      "gallery",
		SortOrder: 10,
		IsActive:  true,
	})
	require.NoError(t, err)

	withImages, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: category.ID,
		Name:       "Photographed tart",
		Slug:       "photographed-tart",
		PriceCents: 500,
		IsActive:   true,
	})
	require.NoError(t, err)

	withoutImages, err := products.Create(ctx, port.CreateProductParams{
		CategoryID: category.ID,
		Name:       "Unphotographed tart",
		Slug:       "unphotographed-tart",
		PriceCents: 500,
		IsActive:   true,
	})
	require.NoError(t, err)

	// Cloudinary delivery URLs carry commas inside the transform segment, which
	// is also the array separator the driver has to get right.
	first := "https://res.cloudinary.com/demo/image/upload/f_auto,q_auto,w_720/boms/products/first.jpg"
	second := "https://res.cloudinary.com/demo/image/upload/f_auto,q_auto,w_720/boms/products/second.jpg"
	require.NoError(t, products.ReplaceProductImages(ctx, withImages.ID, []string{first, second}))

	t.Run("catalog_list_returns_the_gallery_in_sort_order", func(t *testing.T) {
		rows, err := products.CatalogList(ctx, port.CatalogListProductsParams{
			Limit:  20,
			Offset: 0,
		})
		require.NoError(t, err)

		byID := make(map[string][]string, len(rows))
		for _, row := range rows {
			byID[row.ID.String()] = row.ImageURLs
		}

		assert.Equal(t, []string{first, second}, byID[withImages.ID.String()])
		assert.Empty(t, byID[withoutImages.ID.String()],
			"a product with no images returns an empty gallery, never a null array")
	})

	t.Run("manager_list_returns_the_same_gallery", func(t *testing.T) {
		rows, err := products.ManagerList(ctx, port.ManagerListProductsParams{
			Limit:  20,
			Offset: 0,
		})
		require.NoError(t, err)

		byID := make(map[string][]string, len(rows))
		for _, row := range rows {
			byID[row.Product.ID.String()] = row.ImageURLs
		}

		assert.Equal(t, []string{first, second}, byID[withImages.ID.String()])
		assert.Empty(t, byID[withoutImages.ID.String()])
	})

	t.Run("a_soft_deleted_product_leaves_no_gallery_behind", func(t *testing.T) {
		// Its own product: the subtests above assert on theirs, and a fixture one
		// of them deletes would make this suite depend on the order it runs in.
		doomed, err := products.Create(ctx, port.CreateProductParams{
			CategoryID: category.ID,
			Name:       "Withdrawn tart",
			Slug:       "withdrawn-tart",
			PriceCents: 500,
			IsActive:   true,
		})
		require.NoError(t, err)
		require.NoError(t, products.ReplaceProductImages(ctx, doomed.ID, []string{first}))
		require.NoError(t, products.SoftDelete(ctx, doomed.ID))

		rows, err := products.CatalogList(ctx, port.CatalogListProductsParams{
			Limit:  20,
			Offset: 0,
		})
		require.NoError(t, err)

		for _, row := range rows {
			assert.NotEqual(t, doomed.ID, row.ID)
		}
	})
}
