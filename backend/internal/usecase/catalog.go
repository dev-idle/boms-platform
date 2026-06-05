package usecase

import (
	"context"
	"strings"

	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
)

type CatalogUsecase struct {
	categories port.CategoryRepository
	products   port.ProductRepository
}

func NewCatalogUsecase(categories port.CategoryRepository, products port.ProductRepository) *CatalogUsecase {
	return &CatalogUsecase{categories: categories, products: products}
}

func (u *CatalogUsecase) ListCategories(
	ctx context.Context,
	page, pageSize int32,
) ([]dto.CatalogCategoryResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)

	items, err := u.categories.CatalogList(ctx, port.CatalogListCategoriesParams{
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.categories.CatalogListCount(ctx)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	out := make([]dto.CatalogCategoryResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dto.CatalogCategoryResponse{
			ID:        item.ID.String(),
			Name:      item.Name,
			Slug:      item.Slug,
			SortOrder: item.SortOrder,
		})
	}
	return out, total, page, pageSize, nil
}

func (u *CatalogUsecase) ListProducts(
	ctx context.Context,
	page, pageSize int32,
	categoryIDStr string,
) ([]dto.CatalogProductResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)

	var categoryID *uuid.UUID
	if trimmed := strings.TrimSpace(categoryIDStr); trimmed != "" {
		parsed, err := uuid.Parse(trimmed)
		if err != nil {
			return nil, 0, page, pageSize, apperrors.ErrValidation.WithDetail("category_id", "invalid uuid")
		}
		categoryID = &parsed
	}

	items, err := u.products.CatalogList(ctx, port.CatalogListProductsParams{
		CategoryID: categoryID,
		Limit:      pageSize,
		Offset:     utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.products.CatalogListCount(ctx, categoryID)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	out := make([]dto.CatalogProductResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toCatalogProductResponse(item))
	}
	return out, total, page, pageSize, nil
}

func (u *CatalogUsecase) GetProduct(ctx context.Context, id uuid.UUID) (*dto.CatalogProductResponse, error) {
	item, err := u.products.CatalogGetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toCatalogProductResponse(*item)
	return &resp, nil
}

func toCatalogProductResponse(item port.CatalogListProduct) dto.CatalogProductResponse {
	return dto.CatalogProductResponse{
		ID:           item.ID.String(),
		CategoryID:   item.CategoryID.String(),
		CategoryName: item.CategoryName,
		CategorySlug: item.CategorySlug,
		Name:         item.Name,
		Slug:         item.Slug,
		Description:  item.Description,
		PriceCents:   item.PriceCents,
		ImageURL:     item.ImageURL,
	}
}
