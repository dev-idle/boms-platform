package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/boms/backend/internal/config"
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
	tx         port.TxManager
	audit      *auditlogger.Service
	cloudinary config.CloudinaryConfig
	log        *zap.Logger
}

func NewManagerProductUsecase(
	products port.ProductRepository,
	categories port.CategoryRepository,
	tx port.TxManager,
	audit *auditlogger.Service,
	cloudinary config.CloudinaryConfig,
	log *zap.Logger,
) *ManagerProductUsecase {
	return &ManagerProductUsecase{
		products:   products,
		categories: categories,
		tx:         tx,
		audit:      audit,
		cloudinary: cloudinary,
		log:        log,
	}
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
	imageURLs, err := sanitizeManagerProductImageURLs(u.cloudinary, req.ImageURLs)
	if err != nil {
		return nil, err
	}

	var createdID uuid.UUID
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		created, createErr := u.products.Create(txCtx, port.CreateProductParams{
			CategoryID:  categoryID,
			Name:        name,
			Slug:        slug,
			Description: req.Description,
			PriceCents:  req.PriceCents,
			IsAvailable: req.IsAvailable,
		})
		if createErr != nil {
			if errors.Is(createErr, apperrors.ErrConflict) {
				return domainproduct.ErrSlugExists
			}
			return createErr
		}
		createdID = created.ID
		return u.products.ReplaceProductImages(txCtx, created.ID, imageURLs)
	}); err != nil {
		return nil, err
	}

	resp, err := u.managerProductResponse(ctx, createdID)
	if err != nil {
		return nil, err
	}
	u.logAudit(ctx, domaincatalog.AuditActionManagerCreatedProduct, actorID, actorRole, &createdID, "product", nil, toProductAuditFromResponse(resp))
	return resp, nil
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

	productIDs := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		productIDs = append(productIDs, item.Product.ID)
	}
	imagesByProduct, err := u.products.ListProductImagesByProductIDs(ctx, productIDs)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	out := make([]dto.ProductResponse, 0, len(items))
	for _, item := range items {
		out = append(out, *toProductResponse(&item.Product, item.CategoryName, imagesByProduct[item.Product.ID]))
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
	beforeImages, err := u.products.ListProductImagesByProductID(ctx, id)
	if err != nil {
		return nil, err
	}
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
	imageURLs, err := sanitizeManagerProductImageURLs(u.cloudinary, req.ImageURLs)
	if err != nil {
		return nil, err
	}

	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		_, updateErr := u.products.Update(txCtx, port.UpdateProductParams{
			ID:          id,
			CategoryID:  categoryID,
			Name:        name,
			Slug:        slug,
			Description: req.Description,
			PriceCents:  req.PriceCents,
			IsAvailable: req.IsAvailable,
		})
		if updateErr != nil {
			if errors.Is(updateErr, apperrors.ErrConflict) {
				return domainproduct.ErrSlugExists
			}
			if errors.Is(updateErr, apperrors.ErrNotFound) {
				return domainproduct.ErrNotFound
			}
			return updateErr
		}
		return u.products.ReplaceProductImages(txCtx, id, imageURLs)
	}); err != nil {
		return nil, err
	}

	u.logAudit(
		ctx,
		domaincatalog.AuditActionManagerUpdatedProduct,
		actorID,
		actorRole,
		&id,
		"product",
		toProductAudit(before, beforeImages),
		toProductAudit(before, imageURLs),
	)
	return u.managerProductResponse(ctx, id)
}

func (u *ManagerProductUsecase) Delete(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
) error {
	beforeImages, err := u.products.ListProductImagesByProductID(ctx, id)
	if err != nil {
		return err
	}
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

	u.logAudit(ctx, domaincatalog.AuditActionManagerDeletedProduct, actorID, actorRole, &id, "product", toProductAudit(before, beforeImages), nil)
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
	imageURLs, err := u.products.ListProductImagesByProductID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toProductResponse(&item.Product, item.CategoryName, imageURLs), nil
}

func toProductResponse(product *domainproduct.Product, categoryName string, imageURLs []string) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:           product.ID.String(),
		CategoryID:   product.CategoryID.String(),
		CategoryName: categoryName,
		Name:         product.Name,
		Slug:         product.Slug,
		Description:  product.Description,
		PriceCents:   product.PriceCents,
		IsAvailable:  product.IsAvailable,
		ImageURLs:    imageURLs,
		CreatedAt:    product.CreatedAt,
		UpdatedAt:    product.UpdatedAt,
	}
}

func toProductAuditFromResponse(resp *dto.ProductResponse) map[string]any {
	return map[string]any{
		"category_id":  resp.CategoryID,
		"name":         resp.Name,
		"slug":         resp.Slug,
		"price_cents":  resp.PriceCents,
		"is_available": resp.IsAvailable,
		"image_urls":   resp.ImageURLs,
	}
}

func toProductAudit(product *domainproduct.Product, imageURLs []string) map[string]any {
	out := map[string]any{
		"image_urls": imageURLs,
	}
	if product != nil {
		out["category_id"] = product.CategoryID.String()
		out["name"] = product.Name
		out["slug"] = product.Slug
		out["price_cents"] = product.PriceCents
		out["is_available"] = product.IsAvailable
	}
	return out
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
