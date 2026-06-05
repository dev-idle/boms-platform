package postgres

import (
	"context"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaincategory "github.com/boms/backend/internal/domain/category"
	"github.com/boms/backend/internal/port"
	"github.com/google/uuid"
)

type CategoryRepository struct {
	queries *sqlcgen.Queries
}

func NewCategoryRepository(pool *Pool) *CategoryRepository {
	return &CategoryRepository{queries: pool.Queries()}
}

func (r *CategoryRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *CategoryRepository) Create(ctx context.Context, params port.CreateCategoryParams) (*domaincategory.Category, error) {
	row, err := r.q(ctx).CreateCategory(ctx, sqlcgen.CreateCategoryParams{
		Name:      params.Name,
		Slug:      params.Slug,
		SortOrder: params.SortOrder,
		IsActive:  params.IsActive,
	})
	if err != nil {
		return nil, mapRepoError(err, "create category")
	}
	return mapCategory(row), nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domaincategory.Category, error) {
	row, err := r.q(ctx).GetCategoryByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "get category")
	}
	return mapCategory(row), nil
}

func (r *CategoryRepository) Update(ctx context.Context, params port.UpdateCategoryParams) (*domaincategory.Category, error) {
	row, err := r.q(ctx).UpdateCategory(ctx, sqlcgen.UpdateCategoryParams{
		ID:        params.ID,
		Name:      params.Name,
		Slug:      params.Slug,
		SortOrder: params.SortOrder,
		IsActive:  params.IsActive,
	})
	if err != nil {
		return nil, mapRepoError(err, "update category")
	}
	return mapCategory(row), nil
}

func (r *CategoryRepository) SoftDeleteIfNoProducts(ctx context.Context, id uuid.UUID) (int64, error) {
	rows, err := r.q(ctx).SoftDeleteCategoryIfNoProducts(ctx, id)
	if err != nil {
		return 0, mapRepoError(err, "soft delete category if no products")
	}
	return rows, nil
}

func (r *CategoryRepository) ManagerList(ctx context.Context, params port.ManagerListCategoriesParams) ([]domaincategory.Category, error) {
	rows, err := r.q(ctx).ManagerListCategories(ctx, sqlcgen.ManagerListCategoriesParams{
		Limit:  params.Limit,
		Offset: params.Offset,
		Search: optionalSearch(params.Search),
	})
	if err != nil {
		return nil, mapRepoError(err, "manager list categories")
	}
	out := make([]domaincategory.Category, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapCategory(row))
	}
	return out, nil
}

func (r *CategoryRepository) ManagerListCount(ctx context.Context, search *string) (int64, error) {
	count, err := r.q(ctx).ManagerListCategoriesCount(ctx, optionalSearch(search))
	if err != nil {
		return 0, mapRepoError(err, "manager list categories count")
	}
	return count, nil
}

func (r *CategoryRepository) CatalogList(ctx context.Context, params port.CatalogListCategoriesParams) ([]domaincategory.Category, error) {
	rows, err := r.q(ctx).CatalogListCategories(ctx, sqlcgen.CatalogListCategoriesParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "catalog list categories")
	}
	out := make([]domaincategory.Category, 0, len(rows))
	for _, row := range rows {
		out = append(out, domaincategory.Category{
			ID:        row.ID,
			Name:      row.Name,
			Slug:      row.Slug,
			SortOrder: row.SortOrder,
			IsActive:  true,
		})
	}
	return out, nil
}

func (r *CategoryRepository) CatalogListCount(ctx context.Context) (int64, error) {
	count, err := r.q(ctx).CatalogListCategoriesCount(ctx)
	if err != nil {
		return 0, mapRepoError(err, "catalog list categories count")
	}
	return count, nil
}

func mapCategory(row sqlcgen.Category) *domaincategory.Category {
	cat := &domaincategory.Category{
		ID:        row.ID,
		Name:      row.Name,
		Slug:      row.Slug,
		SortOrder: row.SortOrder,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		cat.DeletedAt = &t
	}
	return cat
}

var _ port.CategoryRepository = (*CategoryRepository)(nil)
