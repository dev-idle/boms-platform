package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainproduct "github.com/boms/backend/internal/domain/product"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
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
		CategoryID:      params.CategoryID,
		Name:            params.Name,
		Slug:            params.Slug,
		Description:     params.Description,
		PriceCents:      params.PriceCents,
		IsActive:        params.IsActive,
		LeadTimeMinutes: leadTimeMinutes(params.LeadTime),
		IsCustomizable:  params.IsCustomizable,
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
		ID:              params.ID,
		CategoryID:      params.CategoryID,
		Name:            params.Name,
		Slug:            params.Slug,
		Description:     params.Description,
		PriceCents:      params.PriceCents,
		IsActive:        params.IsActive,
		LeadTimeMinutes: leadTimeMinutes(params.LeadTime),
		IsCustomizable:  params.IsCustomizable,
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
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *ProductRepository) ReplaceProductImages(
	ctx context.Context,
	productID uuid.UUID,
	imageURLs []string,
) error {
	if err := r.q(ctx).DeleteProductImagesByProductID(ctx, productID); err != nil {
		return mapRepoError(err, "delete product images")
	}
	for i, imageURL := range imageURLs {
		if _, err := r.q(ctx).InsertProductImage(ctx, sqlcgen.InsertProductImageParams{
			ProductID: productID,
			SortOrder: int16(i),
			ImageUrl:  imageURL,
		}); err != nil {
			return mapRepoError(err, "insert product image")
		}
	}
	return nil
}

func (r *ProductRepository) ListProductImagesByProductID(
	ctx context.Context,
	productID uuid.UUID,
) ([]string, error) {
	rows, err := r.q(ctx).ListProductImagesByProductID(ctx, productID)
	if err != nil {
		return nil, mapRepoError(err, "list product images")
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ImageUrl)
	}
	return out, nil
}

func (r *ProductRepository) ManagerList(ctx context.Context, params port.ManagerListProductsParams) ([]port.ManagerListProduct, error) {
	rows, err := r.q(ctx).ManagerListProducts(ctx, sqlcgen.ManagerListProductsParams{
		Limit:      params.Limit,
		Offset:     params.Offset,
		CategoryID: params.CategoryID,
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
		CategoryID: categoryID,
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
		CategoryID: params.CategoryID,
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
		CategoryID: categoryID,
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

func (r *ProductRepository) FulfillmentOf(
	ctx context.Context,
	productIDs, comboIDs []uuid.UUID,
	today time.Time,
) (domainorder.Fulfillment, error) {
	row, err := r.q(ctx).GetFulfillmentOf(ctx, sqlcgen.GetFulfillmentOfParams{ProductIds: productIDs, ComboIds: comboIDs, Today: today})
	if err != nil {
		return domainorder.Fulfillment{}, mapRepoError(err, "fulfillment of lines")
	}
	items := domainorder.Fulfillment{
		Kitchen: row.HasKitchenItems,
		Lead:    time.Duration(row.LeadMinutes) * time.Minute,
	}
	if row.SoldOutToday {
		items.SoldOutOn = &today
	}
	return items, nil
}

// StaffList implements port.ProductRepository.
func (r *ProductRepository) StaffList(ctx context.Context, params port.StaffListProductsParams) ([]port.StaffProduct, error) {
	rows, err := r.q(ctx).StaffListProducts(ctx, sqlcgen.StaffListProductsParams{
		SoldOutOn: params.SoldOutOn,
		Limit:     params.Limit,
		Offset:    params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "staff list products")
	}
	out := make([]port.StaffProduct, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.StaffProduct{
			ID:           row.ID,
			Name:         row.Name,
			CategoryName: row.CategoryName,
			Station:      domaincategory.Station(row.Station),
			SoldOutOn:    row.SoldOutOn,
		})
	}
	return out, nil
}

// StaffListCount implements port.ProductRepository.
func (r *ProductRepository) StaffListCount(ctx context.Context, soldOutOn *time.Time) (int64, error) {
	count, err := r.q(ctx).StaffListProductsCount(ctx, soldOutOn)
	if err != nil {
		return 0, mapRepoError(err, "staff list products count")
	}
	return count, nil
}

// SetSoldOut implements port.ProductRepository.
func (r *ProductRepository) SetSoldOut(ctx context.Context, id uuid.UUID, day *time.Time) (*port.StaffProduct, error) {
	row, err := r.q(ctx).SetProductSoldOut(ctx, sqlcgen.SetProductSoldOutParams{ID: id, SoldOutOn: day})
	if err != nil {
		return nil, mapRepoError(err, "set product sold out")
	}
	return &port.StaffProduct{
		ID:           row.ID,
		Name:         row.Name,
		CategoryName: row.CategoryName,
		Station:      domaincategory.Station(row.Station),
		SoldOutOn:    row.SoldOutOn,
	}, nil
}

func (r *ProductRepository) ListOptions(ctx context.Context, productIDs []uuid.UUID) ([]domainproduct.Option, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	rows, err := r.q(ctx).ListProductOptions(ctx, productIDs)
	if err != nil {
		return nil, mapRepoError(err, "list product options")
	}
	out := make([]domainproduct.Option, 0, len(rows))
	for _, row := range rows {
		out = append(out, domainproduct.Option{
			ID:              row.ID,
			ProductID:       row.ProductID,
			Group:           domainproduct.OptionGroup(row.OptionGroup),
			Label:           row.Label,
			PriceDeltaCents: row.PriceDeltaCents,
			SortOrder:       row.SortOrder,
			IsActive:        row.IsActive,
		})
	}
	return out, nil
}

// optionRecord is one option as UpdateProductOptions and CreateProductOptions
// read it from their JSON array.
type optionRecord struct {
	ID              *uuid.UUID `json:"id,omitempty"`
	OptionGroup     string     `json:"option_group"`
	Label           string     `json:"label"`
	PriceDeltaCents int64      `json:"price_delta_cents"`
	SortOrder       int16      `json:"sort_order"`
	IsActive        bool       `json:"is_active"`
}

func (r *ProductRepository) ReplaceOptions(ctx context.Context, productID uuid.UUID, options []domainproduct.Option) error {
	if txFromContext(ctx) == nil {
		return apperrors.Errorf("replace product options: requires a transaction")
	}
	kept := make([]uuid.UUID, 0, len(options))
	var updated, added []optionRecord
	for _, option := range options {
		record := optionRecord{
			OptionGroup:     string(option.Group),
			Label:           option.Label,
			PriceDeltaCents: option.PriceDeltaCents,
			SortOrder:       option.SortOrder,
			IsActive:        option.IsActive,
		}
		if option.ID == uuid.Nil {
			added = append(added, record)
			continue
		}
		id := option.ID
		record.ID = &id
		kept = append(kept, id)
		updated = append(updated, record)
	}
	if err := r.q(ctx).RetireProductOptions(ctx, sqlcgen.RetireProductOptionsParams{ProductID: productID, KeptIds: kept}); err != nil {
		return mapRepoError(err, "retire product options")
	}
	if len(updated) > 0 {
		rows, err := json.Marshal(updated)
		if err != nil {
			return apperrors.Errorf("encode product options: %w", err)
		}
		n, err := r.q(ctx).UpdateProductOptions(ctx, sqlcgen.UpdateProductOptionsParams{ProductID: productID, Rows: rows})
		if err != nil {
			return mapRepoError(err, "update product options")
		}
		if n != int64(len(updated)) {
			return domainproduct.ErrInvalidOption
		}
	}
	if len(added) > 0 {
		rows, err := json.Marshal(added)
		if err != nil {
			return apperrors.Errorf("encode product options: %w", err)
		}
		n, err := r.q(ctx).CreateProductOptions(ctx, sqlcgen.CreateProductOptionsParams{ProductID: productID, Rows: rows})
		if err != nil {
			return mapRepoError(err, "create product options")
		}
		if n != int64(len(added)) {
			return apperrors.Errorf("create product options: inserted %d of %d", n, len(added))
		}
	}
	return nil
}

func mapManagerListProductsRow(row sqlcgen.ManagerListProductsRow) port.ManagerListProduct {
	item := mapManagerJoinedProduct(
		row.ID,
		row.CategoryID,
		row.Name,
		row.Slug,
		row.Description,
		row.PriceCents,
		row.IsActive,
		row.CreatedAt,
		row.UpdatedAt,
		row.DeletedAt,
		row.LeadTimeMinutes,
		row.IsCustomizable,
		row.CategoryName,
	)
	item.ImageURLs = row.ImageUrls
	return item
}

func mapManagerGetProductRow(row sqlcgen.ManagerGetProductByIDRow) port.ManagerListProduct {
	return mapManagerJoinedProduct(
		row.ID,
		row.CategoryID,
		row.Name,
		row.Slug,
		row.Description,
		row.PriceCents,
		row.IsActive,
		row.CreatedAt,
		row.UpdatedAt,
		row.DeletedAt,
		row.LeadTimeMinutes,
		row.IsCustomizable,
		row.CategoryName,
	)
}

func mapManagerJoinedProduct(
	id uuid.UUID,
	categoryID uuid.UUID,
	name string,
	slug string,
	description *string,
	priceCents int64,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
	leadMinutes int32,
	isCustomizable bool,
	categoryName string,
) port.ManagerListProduct {
	return port.ManagerListProduct{
		Product: domainproduct.Product{
			ID:             id,
			CategoryID:     categoryID,
			Name:           name,
			Slug:           slug,
			Description:    description,
			PriceCents:     priceCents,
			IsActive:       isActive,
			CreatedAt:      createdAt,
			UpdatedAt:      updatedAt,
			DeletedAt:      deletedAt,
			LeadTime:       time.Duration(leadMinutes) * time.Minute,
			IsCustomizable: isCustomizable,
		},
		CategoryName: categoryName,
	}
}

func mapProduct(row sqlcgen.Product) *domainproduct.Product {
	return &domainproduct.Product{
		ID:             row.ID,
		CategoryID:     row.CategoryID,
		Name:           row.Name,
		Slug:           row.Slug,
		PriceCents:     row.PriceCents,
		IsActive:       row.IsActive,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
		Description:    row.Description,
		DeletedAt:      row.DeletedAt,
		LeadTime:       time.Duration(row.LeadTimeMinutes) * time.Minute,
		IsCustomizable: row.IsCustomizable,
	}
}

// leadTimeMinutes stores a lead time the domain has bounded to a week, so the
// minutes always fit.
func leadTimeMinutes(d time.Duration) int32 {
	return utils.Int32FromInt64(int64(d / time.Minute))
}

func mapCatalogListProductsRow(row sqlcgen.CatalogListProductsRow) port.CatalogListProduct {
	item := mapCatalogProductFields(
		row.ID,
		row.CategoryID,
		row.Name,
		row.Slug,
		row.Description,
		row.PriceCents,
		row.IsCustomizable,
		row.SoldOutOn,
		row.CategoryName,
		row.CategorySlug,
	)
	item.ImageURLs = row.ImageUrls
	return item
}

func mapCatalogGetProductRow(row sqlcgen.CatalogGetProductByIDRow) port.CatalogListProduct {
	return mapCatalogProductFields(
		row.ID,
		row.CategoryID,
		row.Name,
		row.Slug,
		row.Description,
		row.PriceCents,
		row.IsCustomizable,
		row.SoldOutOn,
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
		row.IsCustomizable,
		row.SoldOutOn,
		row.CategoryName,
		row.CategorySlug,
	)
}

func mapCatalogProductFields(
	id uuid.UUID,
	categoryID uuid.UUID,
	name string,
	slug string,
	description *string,
	priceCents int64,
	isCustomizable bool,
	soldOutOn *time.Time,
	categoryName string,
	categorySlug string,
) port.CatalogListProduct {
	return port.CatalogListProduct{
		ID:             id,
		CategoryID:     categoryID,
		Name:           name,
		Slug:           slug,
		PriceCents:     priceCents,
		IsCustomizable: isCustomizable,
		SoldOutOn:      soldOutOn,
		CategoryName:   categoryName,
		CategorySlug:   categorySlug,
		Description:    description,
	}
}

var _ port.ProductRepository = (*ProductRepository)(nil)
