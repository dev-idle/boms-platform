package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainsaved "github.com/boms/backend/internal/domain/saved"
	"github.com/boms/backend/internal/port"
)

// SavedProductRepository keeps the products customers keep on their lists.
type SavedProductRepository struct {
	queries *sqlcgen.Queries
}

func NewSavedProductRepository(pool *Pool) *SavedProductRepository {
	return &SavedProductRepository{queries: pool.Queries()}
}

func (r *SavedProductRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// Save implements port.SavedProductRepository.
func (r *SavedProductRepository) Save(ctx context.Context, userID, productID uuid.UUID, list domainsaved.List) (bool, error) {
	n, err := r.q(ctx).SaveProduct(ctx, sqlcgen.SaveProductParams{
		UserID:    userID,
		ProductID: productID,
		List:      sqlcgen.SavedList(list),
	})
	if err != nil {
		return false, mapRepoError(err, "save product")
	}
	return n > 0, nil
}

// CountOthers implements port.SavedProductRepository.
func (r *SavedProductRepository) CountOthers(ctx context.Context, userID, productID uuid.UUID, list domainsaved.List) (int64, error) {
	n, err := r.q(ctx).CountOtherSavedProducts(ctx, sqlcgen.CountOtherSavedProductsParams{
		UserID:    userID,
		ProductID: productID,
		List:      sqlcgen.SavedList(list),
	})
	if err != nil {
		return 0, mapRepoError(err, "count saved products")
	}
	return n, nil
}

// Remove implements port.SavedProductRepository.
func (r *SavedProductRepository) Remove(ctx context.Context, userID, productID uuid.UUID, list domainsaved.List) error {
	if err := r.q(ctx).RemoveSavedProduct(ctx, sqlcgen.RemoveSavedProductParams{
		UserID:    userID,
		ProductID: productID,
		List:      sqlcgen.SavedList(list),
	}); err != nil {
		return mapRepoError(err, "remove saved product")
	}
	return nil
}

// ListByUser implements port.SavedProductRepository.
func (r *SavedProductRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]port.SavedProduct, error) {
	rows, err := r.q(ctx).ListSavedProducts(ctx, userID)
	if err != nil {
		return nil, mapRepoError(err, "list saved products")
	}
	out := make([]port.SavedProduct, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.SavedProduct{
			List:    domainsaved.List(row.List),
			SavedAt: row.SavedAt,
			Product: port.CatalogListProduct{
				ID:             row.ID,
				CategoryID:     row.CategoryID,
				Name:           row.Name,
				Slug:           row.Slug,
				PriceCents:     row.PriceCents,
				IsCustomizable: row.IsCustomizable,
				SoldOutOn:      row.SoldOutOn,
				CategoryName:   row.CategoryName,
				CategorySlug:   row.CategorySlug,
			},
		})
	}
	return out, nil
}

// ListForExport implements port.SavedProductRepository.
func (r *SavedProductRepository) ListForExport(ctx context.Context, userID uuid.UUID) ([]port.SavedProductEntry, error) {
	rows, err := r.q(ctx).ListSavedProductsForExport(ctx, userID)
	if err != nil {
		return nil, mapRepoError(err, "list saved products for export")
	}
	out := make([]port.SavedProductEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.SavedProductEntry{
			List:        domainsaved.List(row.List),
			SavedAt:     row.SavedAt,
			ProductID:   row.ID,
			ProductName: row.Name,
		})
	}
	return out, nil
}

// RemoveAll implements port.SavedProductRepository.
func (r *SavedProductRepository) RemoveAll(ctx context.Context, userID uuid.UUID) error {
	if err := r.q(ctx).RemoveAllSavedProducts(ctx, userID); err != nil {
		return mapRepoError(err, "remove all saved products")
	}
	return nil
}

var _ port.SavedProductRepository = (*SavedProductRepository)(nil)
