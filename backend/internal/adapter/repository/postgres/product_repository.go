package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainproduct "github.com/boms/backend/internal/domain/product"
	"github.com/boms/backend/internal/port"
	"github.com/google/uuid"
)

type ProductRepository struct {
	queries *sqlcgen.Queries
}

func NewProductRepository(pool *Pool) *ProductRepository {
	return &ProductRepository{queries: pool.Queries()}
}

func (r *ProductRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *ProductRepository) Create(ctx context.Context, params port.CreateProductParams) (*domainproduct.Product, error) {
	row, err := r.q(ctx).CreateProduct(ctx, sqlcgen.CreateProductParams{
		CategoryID:  params.CategoryID,
		Name:        params.Name,
		Slug:        params.Slug,
		Description: optionalString(params.Description),
		PriceCents:  params.PriceCents,
		IsAvailable: params.IsAvailable,
		ImageUrl:    optionalString(params.ImageURL),
	})
	if err != nil {
		return nil, mapRepoError(err, "create product")
	}
	return mapProduct(row), nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainproduct.Product, error) {
	row, err := r.q(ctx).GetProductByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "get product")
	}
	return mapProduct(row), nil
}

func (r *ProductRepository) Update(ctx context.Context, params port.UpdateProductParams) (*domainproduct.Product, error) {
	row, err := r.q(ctx).UpdateProduct(ctx, sqlcgen.UpdateProductParams{
		ID:          params.ID,
		CategoryID:  params.CategoryID,
		Name:        params.Name,
		Slug:        params.Slug,
		Description: optionalString(params.Description),
		PriceCents:  params.PriceCents,
		IsAvailable: params.IsAvailable,
		ImageUrl:    optionalString(params.ImageURL),
	})
	if err != nil {
		return nil, mapRepoError(err, "update product")
	}
	return mapProduct(row), nil
}

func (r *ProductRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q(ctx).SoftDeleteProduct(ctx, id)
	if err != nil {
		return mapRepoError(err, "soft delete product")
	}
	if rows == 0 {
		return mapRepoError(sql.ErrNoRows, "soft delete product")
	}
	return nil
}

func (r *ProductRepository) ManagerList(ctx context.Context, params port.ManagerListProductsParams) ([]port.ManagerListProduct, error) {
	rows, err := r.q(ctx).ManagerListProducts(ctx, sqlcgen.ManagerListProductsParams{
		Limit:      params.Limit,
		Offset:     params.Offset,
		CategoryID: optionalUUID(params.CategoryID),
		Search:     optionalSearch(params.Search),
	})
	if err != nil {
		return nil, mapRepoError(err, "manager list products")
	}
	out := make([]port.ManagerListProduct, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapManagerListProductsRow(row))
	}
	return out, nil
}

func (r *ProductRepository) ManagerListCount(ctx context.Context, categoryID *uuid.UUID, search *string) (int64, error) {
	count, err := r.q(ctx).ManagerListProductsCount(ctx, sqlcgen.ManagerListProductsCountParams{
		CategoryID: optionalUUID(categoryID),
		Search:     optionalSearch(search),
	})
	if err != nil {
		return 0, mapRepoError(err, "manager list products count")
	}
	return count, nil
}

func (r *ProductRepository) ManagerGetByID(ctx context.Context, id uuid.UUID) (*port.ManagerListProduct, error) {
	row, err := r.q(ctx).ManagerGetProductByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "manager get product")
	}
	out := mapManagerGetProductRow(row)
	return &out, nil
}

func (r *ProductRepository) CatalogList(ctx context.Context, params port.CatalogListProductsParams) ([]port.CatalogListProduct, error) {
	rows, err := r.q(ctx).CatalogListProducts(ctx, sqlcgen.CatalogListProductsParams{
		Limit:      params.Limit,
		Offset:     params.Offset,
		CategoryID: optionalUUID(params.CategoryID),
		Search:     optionalSearch(params.Search),
	})
	if err != nil {
		return nil, mapRepoError(err, "catalog list products")
	}
	out := make([]port.CatalogListProduct, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCatalogListProductsRow(row))
	}
	return out, nil
}

func (r *ProductRepository) CatalogListCount(
	ctx context.Context,
	categoryID *uuid.UUID,
	search *string,
) (int64, error) {
	count, err := r.q(ctx).CatalogListProductsCount(ctx, sqlcgen.CatalogListProductsCountParams{
		CategoryID: optionalUUID(categoryID),
		Search:     optionalSearch(search),
	})
	if err != nil {
		return 0, mapRepoError(err, "catalog list products count")
	}
	return count, nil
}

func (r *ProductRepository) CatalogGetByID(ctx context.Context, id uuid.UUID) (*port.CatalogListProduct, error) {
	row, err := r.q(ctx).CatalogGetProductByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "catalog get product")
	}
	out := mapCatalogGetProductRow(row)
	return &out, nil
}

func (r *ProductRepository) CatalogGetByIDs(ctx context.Context, ids []uuid.UUID) ([]port.CatalogListProduct, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.q(ctx).CatalogGetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, mapRepoError(err, "catalog get products by ids")
	}
	out := make([]port.CatalogListProduct, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCatalogGetProductsByIDsRow(row))
	}
	return out, nil
}

func mapManagerListProductsRow(row sqlcgen.ManagerListProductsRow) port.ManagerListProduct {
	return mapManagerJoinedProduct(
		row.ID,
		row.CategoryID,
		row.Name,
		row.Slug,
		row.Description,
		row.PriceCents,
		row.IsAvailable,
		row.ImageUrl,
		row.CreatedAt,
		row.UpdatedAt,
		row.DeletedAt,
		row.CategoryName,
	)
}

func mapManagerGetProductRow(row sqlcgen.ManagerGetProductByIDRow) port.ManagerListProduct {
	return mapManagerJoinedProduct(
		row.ID,
		row.CategoryID,
		row.Name,
		row.Slug,
		row.Description,
		row.PriceCents,
		row.IsAvailable,
		row.ImageUrl,
		row.CreatedAt,
		row.UpdatedAt,
		row.DeletedAt,
		row.CategoryName,
	)
}

func mapManagerJoinedProduct(
	id uuid.UUID,
	categoryID uuid.UUID,
	name string,
	slug string,
	description sql.NullString,
	priceCents int64,
	isAvailable bool,
	imageURL sql.NullString,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt sql.NullTime,
	categoryName string,
) port.ManagerListProduct {
	p := &domainproduct.Product{
		ID:          id,
		CategoryID:  categoryID,
		Name:        name,
		Slug:        slug,
		PriceCents:  priceCents,
		IsAvailable: isAvailable,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
	if description.Valid {
		desc := description.String
		p.Description = &desc
	}
	if imageURL.Valid {
		url := imageURL.String
		p.ImageURL = &url
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		p.DeletedAt = &t
	}
	return port.ManagerListProduct{
		Product:      *p,
		CategoryName: categoryName,
	}
}

func mapProduct(row sqlcgen.Product) *domainproduct.Product {
	p := &domainproduct.Product{
		ID:          row.ID,
		CategoryID:  row.CategoryID,
		Name:        row.Name,
		Slug:        row.Slug,
		PriceCents:  row.PriceCents,
		IsAvailable: row.IsAvailable,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.Description.Valid {
		desc := row.Description.String
		p.Description = &desc
	}
	if row.ImageUrl.Valid {
		url := row.ImageUrl.String
		p.ImageURL = &url
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		p.DeletedAt = &t
	}
	return p
}

func mapCatalogListProductsRow(row sqlcgen.CatalogListProductsRow) port.CatalogListProduct {
	return mapCatalogProductFields(
		row.ID,
		row.CategoryID,
		row.Name,
		row.Slug,
		row.Description,
		row.PriceCents,
		row.ImageUrl,
		row.CategoryName,
		row.CategorySlug,
	)
}

func mapCatalogGetProductRow(row sqlcgen.CatalogGetProductByIDRow) port.CatalogListProduct {
	return mapCatalogProductFields(
		row.ID,
		row.CategoryID,
		row.Name,
		row.Slug,
		row.Description,
		row.PriceCents,
		row.ImageUrl,
		row.CategoryName,
		row.CategorySlug,
	)
}

func mapCatalogGetProductsByIDsRow(row sqlcgen.CatalogGetProductsByIDsRow) port.CatalogListProduct {
	return mapCatalogProductFields(
		row.ID,
		row.CategoryID,
		row.Name,
		row.Slug,
		row.Description,
		row.PriceCents,
		row.ImageUrl,
		row.CategoryName,
		row.CategorySlug,
	)
}

func mapCatalogProductFields(
	id uuid.UUID,
	categoryID uuid.UUID,
	name string,
	slug string,
	description sql.NullString,
	priceCents int64,
	imageURL sql.NullString,
	categoryName string,
	categorySlug string,
) port.CatalogListProduct {
	out := port.CatalogListProduct{
		ID:           id,
		CategoryID:   categoryID,
		Name:         name,
		Slug:         slug,
		PriceCents:   priceCents,
		CategoryName: categoryName,
		CategorySlug: categorySlug,
	}
	if description.Valid {
		desc := description.String
		out.Description = &desc
	}
	if imageURL.Valid {
		url := imageURL.String
		out.ImageURL = &url
	}
	return out
}

var _ port.ProductRepository = (*ProductRepository)(nil)
