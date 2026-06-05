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
	combos     port.ComboRepository
}

func NewCatalogUsecase(
	categories port.CategoryRepository,
	products port.ProductRepository,
	combos port.ComboRepository,
) *CatalogUsecase {
	return &CatalogUsecase{categories: categories, products: products, combos: combos}
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

func (u *CatalogUsecase) ListCombos(
	ctx context.Context,
	page, pageSize int32,
) ([]dto.CatalogComboResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)

	items, err := u.combos.CatalogList(ctx, port.CatalogListCombosParams{
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.combos.CatalogListCount(ctx)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	out := make([]dto.CatalogComboResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toCatalogComboResponse(item))
	}
	return out, total, page, pageSize, nil
}

func (u *CatalogUsecase) GetCombo(ctx context.Context, id uuid.UUID) (*dto.CatalogComboResponse, error) {
	item, err := u.combos.CatalogGetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toCatalogComboResponse(*item)
	return &resp, nil
}

func toCatalogComboResponse(item port.CatalogCombo) dto.CatalogComboResponse {
	return dto.CatalogComboResponse{
		ID:         item.ID.String(),
		Name:       item.Name,
		Slug:       item.Slug,
		PriceCents: item.PriceCents,
		StartsAt:   item.StartsAt,
		EndsAt:     item.EndsAt,
		Items:      toComboItemResponses(item.Items),
	}
}
