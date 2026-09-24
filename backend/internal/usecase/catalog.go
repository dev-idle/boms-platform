package usecase

import (
	"context"
	"strings"

	domaincategory "github.com/boms/backend/internal/domain/category"
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

	items, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]domaincategory.Category, error) {
			return u.categories.CatalogList(ctx, port.CatalogListCategoriesParams{
				Limit:  pageSize,
				Offset: utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.categories.CatalogListCount(ctx)
		},
	)
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
	categoryIDStr, search string,
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

	var searchPtr *string
	if trimmed := strings.TrimSpace(search); trimmed != "" {
		if len(trimmed) > catalogSearchMaxLen {
			return nil, 0, page, pageSize, apperrors.ErrValidation.WithDetail("search", "must be at most 100 characters")
		}
		searchPtr = &trimmed
	}

	items, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.CatalogListProduct, error) {
			return u.products.CatalogList(ctx, port.CatalogListProductsParams{
				CategoryID: categoryID,
				Search:     searchPtr,
				Limit:      pageSize,
				Offset:     utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.products.CatalogListCount(ctx, categoryID, searchPtr)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	// The gallery arrives with each row, so the page is one round trip.
	out := make([]dto.CatalogProductResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toCatalogProductResponse(item, item.ImageURLs))
	}
	return out, total, page, pageSize, nil
}

func (u *CatalogUsecase) GetProduct(ctx context.Context, id uuid.UUID) (*dto.CatalogProductResponse, error) {
	item, err := u.products.CatalogGetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	imageURLs, err := u.products.ListProductImagesByProductID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toCatalogProductResponse(*item, imageURLs)
	return &resp, nil
}

func toCatalogProductResponse(item port.CatalogListProduct, imageURLs []string) dto.CatalogProductResponse {
	return dto.CatalogProductResponse{
		ID:           item.ID.String(),
		CategoryID:   item.CategoryID.String(),
		CategoryName: item.CategoryName,
		CategorySlug: item.CategorySlug,
		Name:         item.Name,
		Slug:         item.Slug,
		Description:  item.Description,
		PriceCents:   item.PriceCents,
		ImageURLs:    imageURLs,
	}
}

func (u *CatalogUsecase) ListCombos(
	ctx context.Context,
	page, pageSize int32,
) ([]dto.CatalogComboResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)

	items, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.CatalogCombo, error) {
			return u.combos.CatalogList(ctx, port.CatalogListCombosParams{
				Limit:  pageSize,
				Offset: utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.combos.CatalogListCount(ctx)
		},
	)
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
		ImageURL:   item.ImageURL,
		StartsAt:   item.StartsAt,
		EndsAt:     item.EndsAt,
		Items:      toComboItemResponses(item.Items),
	}
}
