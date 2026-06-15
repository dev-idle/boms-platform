package usecase

import (
	"context"
	"errors"
	"strings"

	domaincatalog "github.com/boms/backend/internal/domain/catalog"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainproduct "github.com/boms/backend/internal/domain/product"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ManagerProductUsecase struct {
	products   port.ProductRepository
	categories port.CategoryRepository
	audit      *auditlogger.Service
	log        *zap.Logger
}

func NewManagerProductUsecase(
	products port.ProductRepository,
	categories port.CategoryRepository,
	audit *auditlogger.Service,
	log *zap.Logger,
) *ManagerProductUsecase {
	return &ManagerProductUsecase{products: products, categories: categories, audit: audit, log: log}
}

func (u *ManagerProductUsecase) Create(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.CreateProductRequest,
) (*dto.ProductResponse, error) {
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("category_id", "invalid uuid")
	}
	if err := u.ensureCategoryExists(ctx, categoryID); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.ErrValidation.WithDetail("name", "required")
	}
	slug, err := resolveManagerCatalogSlug(name, req.Slug, true)
	if err != nil {
		return nil, err
	}

	created, err := u.products.Create(ctx, port.CreateProductParams{
		CategoryID:  categoryID,
		Name:        name,
		Slug:        slug,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		IsAvailable: req.IsAvailable,
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domainproduct.ErrSlugExists
		}
		return nil, err
	}

	u.logAudit(ctx, domaincatalog.AuditActionManagerCreatedProduct, actorID, actorRole, &created.ID, "product", nil, toProductAudit(created))
	return u.managerProductResponse(ctx, created.ID)
}

func (u *ManagerProductUsecase) Get(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error) {
	return u.managerProductResponse(ctx, id)
}

func (u *ManagerProductUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	categoryIDStr, search string,
) ([]dto.ProductResponse, int64, int32, int32, error) {
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
		searchPtr = &trimmed
	}

	items, err := u.products.ManagerList(ctx, port.ManagerListProductsParams{
		CategoryID: categoryID,
		Search:     searchPtr,
		Limit:      pageSize,
		Offset:     utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.products.ManagerListCount(ctx, categoryID, searchPtr)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	out := make([]dto.ProductResponse, 0, len(items))
	for _, item := range items {
		out = append(out, *toProductResponse(&item.Product, item.CategoryName))
	}
	return out, total, page, pageSize, nil
}

func (u *ManagerProductUsecase) Update(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
	req dto.UpdateProductRequest,
) (*dto.ProductResponse, error) {
	before, err := u.products.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainproduct.ErrNotFound
		}
		return nil, err
	}

	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, apperrors.ErrValidation.WithDetail("category_id", "invalid uuid")
	}
	if err := u.ensureCategoryExists(ctx, categoryID); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.ErrValidation.WithDetail("name", "required")
	}
	slug, err := resolveManagerCatalogSlug(name, req.Slug, false)
	if err != nil {
		return nil, err
	}

	updated, err := u.products.Update(ctx, port.UpdateProductParams{
		ID:          id,
		CategoryID:  categoryID,
		Name:        name,
		Slug:        slug,
		Description: req.Description,
		PriceCents:  req.PriceCents,
		IsAvailable: req.IsAvailable,
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domainproduct.ErrSlugExists
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainproduct.ErrNotFound
		}
		return nil, err
	}

	u.logAudit(ctx, domaincatalog.AuditActionManagerUpdatedProduct, actorID, actorRole, &id, "product", toProductAudit(before), toProductAudit(updated))
	return u.managerProductResponse(ctx, id)
}

func (u *ManagerProductUsecase) Delete(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
) error {
	before, err := u.products.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return domainproduct.ErrNotFound
		}
		return err
	}

	if err := u.products.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return domainproduct.ErrNotFound
		}
		return err
	}

	u.logAudit(ctx, domaincatalog.AuditActionManagerDeletedProduct, actorID, actorRole, &id, "product", toProductAudit(before), nil)
	return nil
}

func (u *ManagerProductUsecase) ensureCategoryExists(ctx context.Context, categoryID uuid.UUID) error {
	cat, err := u.categories.GetByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return domaincategory.ErrNotFound
		}
		return err
	}
	if !cat.IsActive {
		return domaincategory.ErrInactive
	}
	return nil
}

func (u *ManagerProductUsecase) managerProductResponse(ctx context.Context, id uuid.UUID) (*dto.ProductResponse, error) {
	item, err := u.products.ManagerGetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainproduct.ErrNotFound
		}
		return nil, err
	}
	return toProductResponse(&item.Product, item.CategoryName), nil
}

func toProductResponse(product *domainproduct.Product, categoryName string) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:           product.ID.String(),
		CategoryID:   product.CategoryID.String(),
		CategoryName: categoryName,
		Name:         product.Name,
		Slug:         product.Slug,
		Description:  product.Description,
		PriceCents:   product.PriceCents,
		IsAvailable:  product.IsAvailable,
		ImageURL:     product.ImageURL,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
	}
}

func toProductAudit(product *domainproduct.Product) map[string]any {
	return map[string]any{
		"category_id":  product.CategoryID.String(),
		"name":         product.Name,
		"slug":         product.Slug,
		"price_cents":  product.PriceCents,
		"is_available": product.IsAvailable,
	}
}

func (u *ManagerProductUsecase) logAudit(
	ctx context.Context,
	action domainuser.AuditAction,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	targetID *uuid.UUID,
	targetType string,
	before, after any,
) {
	recordAudit(u.log, u.audit, ctx, action, actorID, actorRole, targetID, targetType, before, after)
}
