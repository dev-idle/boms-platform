package bootstrap

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/boms/backend/internal/config"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/google/uuid"
)

// seedListLimit is wide enough to hold the whole seed set in one page.
const seedListLimit int32 = 500

// CatalogSeedDeps are the repositories the catalog seed writes through.
type CatalogSeedDeps struct {
	Categories port.CategoryRepository
	Products   port.ProductRepository
	Combos     port.ComboRepository
	Discounts  port.DiscountCodeRepository
}

// CatalogSeedReport counts what the run created; existing rows are left untouched.
type CatalogSeedReport struct {
	Categories    int
	Products      int
	Combos        int
	DiscountCodes int
}

// SeedCatalog fills an empty development database with the demo catalog. It is idempotent:
// rows are matched by slug or code, so a second run creates nothing.
func SeedCatalog(
	ctx context.Context,
	cfg *config.Config,
	deps CatalogSeedDeps,
	now time.Time,
) (CatalogSeedReport, error) {
	if cfg == nil || strings.ToLower(cfg.App.Env) != "development" {
		return CatalogSeedReport{}, errors.New("catalog seed runs in development only")
	}

	report := CatalogSeedReport{}

	categoryIDs, created, err := seedCategoriesInto(ctx, deps.Categories)
	if err != nil {
		return report, err
	}
	report.Categories = created

	products, created, err := seedProductsInto(ctx, deps.Products, categoryIDs)
	if err != nil {
		return report, err
	}
	report.Products = created

	if report.Combos, err = seedCombosInto(ctx, deps.Combos, products, now); err != nil {
		return report, err
	}
	if report.DiscountCodes, err = seedDiscountsInto(ctx, deps.Discounts, now); err != nil {
		return report, err
	}
	return report, nil
}

func seedCategoriesInto(
	ctx context.Context,
	repo port.CategoryRepository,
) (map[string]uuid.UUID, int, error) {
	existing, err := repo.ManagerList(ctx, port.ManagerListCategoriesParams{Limit: seedListLimit})
	if err != nil {
		return nil, 0, err
	}
	ids := make(map[string]uuid.UUID, len(seedCategories))
	for _, category := range existing {
		ids[strings.ToLower(category.Slug)] = category.ID
	}

	created := 0
	for _, seed := range seedCategories {
		if _, ok := ids[seed.Slug]; ok {
			continue
		}
		category, err := repo.Create(ctx, port.CreateCategoryParams{
			Name:      seed.Name,
			Slug:      seed.Slug,
			SortOrder: seed.SortOrder,
			IsActive:  seed.IsActive,
		})
		if err != nil {
			return nil, created, err
		}
		ids[seed.Slug] = category.ID
		created++
	}
	return ids, created, nil
}

// seededProduct is what later stages need from a product: its id and unit price.
type seededProduct struct {
	ID         uuid.UUID
	PriceCents int64
}

func seedProductsInto(
	ctx context.Context,
	repo port.ProductRepository,
	categoryIDs map[string]uuid.UUID,
) (map[string]seededProduct, int, error) {
	existing, err := repo.ManagerList(ctx, port.ManagerListProductsParams{Limit: seedListLimit})
	if err != nil {
		return nil, 0, err
	}
	products := make(map[string]seededProduct, len(seedProducts))
	for _, row := range existing {
		products[strings.ToLower(row.Product.Slug)] = seededProduct{
			ID:         row.Product.ID,
			PriceCents: row.Product.PriceCents,
		}
	}

	created := 0
	for _, seed := range seedProducts {
		if _, ok := products[seed.Slug]; ok {
			continue
		}
		categoryID, ok := categoryIDs[seed.CategorySlug]
		if !ok {
			return nil, created, errors.New("catalog seed: unknown category " + seed.CategorySlug)
		}
		description := seed.Description
		product, err := repo.Create(ctx, port.CreateProductParams{
			CategoryID:  categoryID,
			Name:        seed.Name,
			Slug:        seed.Slug,
			Description: &description,
			PriceCents:  seed.PriceCents,
			IsActive:    seed.IsActive,
		})
		if err != nil {
			return nil, created, err
		}
		products[seed.Slug] = seededProduct{ID: product.ID, PriceCents: product.PriceCents}
		created++
	}
	return products, created, nil
}

func seedCombosInto(
	ctx context.Context,
	repo port.ComboRepository,
	products map[string]seededProduct,
	now time.Time,
) (int, error) {
	existing, err := repo.ManagerList(ctx, port.ManagerListCombosParams{Limit: seedListLimit})
	if err != nil {
		return 0, err
	}
	known := make(map[string]struct{}, len(existing))
	for _, combo := range existing {
		known[strings.ToLower(combo.Slug)] = struct{}{}
	}

	created := 0
	for _, seed := range seedCombos {
		if _, ok := known[seed.Slug]; ok {
			continue
		}
		items := make([]port.ComboItemParams, 0, len(seed.Items))
		var sumCents int64
		for _, item := range seed.Items {
			product, ok := products[item.ProductSlug]
			if !ok {
				return created, errors.New("catalog seed: unknown product " + item.ProductSlug)
			}
			items = append(items, port.ComboItemParams{ProductID: product.ID, Quantity: item.Quantity})
			sumCents += product.PriceCents * int64(item.Quantity)
		}

		startsAt, endsAt := windowRange(seed.Window, now)
		combo, err := repo.Create(ctx, port.CreateComboParams{
			Name:       seed.Name,
			Slug:       seed.Slug,
			PriceCents: bundlePriceCents(sumCents, seed.DiscountPercent),
			StartsAt:   startsAt,
			EndsAt:     endsAt,
			IsActive:   seed.IsActive,
		})
		if err != nil {
			return created, err
		}
		if err := repo.ReplaceItems(ctx, combo.ID, items); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

func seedDiscountsInto(
	ctx context.Context,
	repo port.DiscountCodeRepository,
	now time.Time,
) (int, error) {
	created := 0
	for _, seed := range seedDiscountCodes {
		if _, err := repo.GetByCode(ctx, seed.Code); err == nil {
			continue
		} else if !errors.Is(err, apperrors.ErrNotFound) {
			return created, err
		}

		startsAt, endsAt := windowRange(seed.Window, now)
		code, err := repo.Create(ctx, port.CreateDiscountCodeParams{
			Code:             seed.Code,
			DiscountType:     seed.DiscountType,
			Value:            seed.Value,
			MinOrderCents:    seed.MinOrderCents,
			MaxUses:          seed.MaxUses,
			MaxDiscountCents: seed.MaxDiscountCents,
			StartsAt:         startsAt,
			EndsAt:           endsAt,
			IsActive:         seed.IsActive,
		})
		if err != nil {
			return created, err
		}
		for range seed.UsedCount {
			if _, err := repo.IncrementUsedCount(ctx, code.ID); err != nil {
				return created, err
			}
		}
		created++
	}
	return created, nil
}

// maxBundleDiscountPercent is the house ceiling for a combo saving.
const maxBundleDiscountPercent int64 = 8

// bundlePriceCents prices a combo below the sum of its items and rounds down to a whole dime
// for a clean shelf price — unless that rounding would push the saving past the house ceiling.
func bundlePriceCents(sumCents, discountPercent int64) int64 {
	discounted := sumCents - (sumCents*discountPercent)/100
	rounded := (discounted / 10) * 10
	floor := sumCents - (sumCents*maxBundleDiscountPercent)/100
	if rounded < floor {
		return discounted
	}
	return rounded
}

// windowRange turns a seed window into a concrete range around the run time.
func windowRange(window seedWindow, now time.Time) (time.Time, time.Time) {
	switch window {
	case windowUpcoming:
		return now.AddDate(0, 0, 5), now.AddDate(0, 0, 35)
	case windowClosed:
		return now.AddDate(0, 0, -45), now.AddDate(0, 0, -7)
	case windowLive:
		return now.AddDate(0, 0, -7), now.AddDate(0, 0, 30)
	default:
		return now.AddDate(0, 0, -7), now.AddDate(0, 0, 30)
	}
}
