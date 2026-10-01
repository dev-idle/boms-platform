package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

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
	leadTime, err := productLeadTime(req.LeadTimeMinutes)
	if err != nil {
		return nil, err
	}
	imageURLs, err := sanitizeManagerProductImageURLs(u.cloudinary, req.ImageURLs)
	if err != nil {
		return nil, err
	}
	options, err := productOptions(req.Options)
	if err != nil {
		return nil, err
	}

	var createdID uuid.UUID
	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		created, createErr := u.products.Create(txCtx, port.CreateProductParams{
			CategoryID:     categoryID,
			Name:           name,
			Slug:           slug,
			Description:    req.Description,
			PriceCents:     req.PriceCents,
			IsActive:       req.IsActive,
			LeadTime:       leadTime,
			IsCustomizable: req.IsCustomizable,
		})
		if createErr != nil {
			if errors.Is(createErr, apperrors.ErrConflict) {
				return domainproduct.ErrSlugExists
			}
			return createErr
		}
		createdID = created.ID
		if len(options) > 0 {
			if err := u.products.ReplaceOptions(txCtx, created.ID, options); err != nil {
				return err
			}
		}
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

	items, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.ManagerListProduct, error) {
			return u.products.ManagerList(ctx, port.ManagerListProductsParams{
				CategoryID: categoryID,
				Search:     searchPtr,
				Limit:      pageSize,
				Offset:     utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.products.ManagerListCount(ctx, categoryID, searchPtr)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	out := make([]dto.ProductResponse, 0, len(items))
	for _, item := range items {
		out = append(out, *toProductResponse(&item.Product, item.CategoryName, item.ImageURLs))
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
	beforeOptions, err := u.products.ListOptions(ctx, []uuid.UUID{id})
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
	leadTime, err := productLeadTime(req.LeadTimeMinutes)
	if err != nil {
		return nil, err
	}
	var imageURLs []string
	replaceImages := req.ImageURLs != nil
	if replaceImages {
		imageURLs, err = sanitizeManagerProductImageURLs(u.cloudinary, *req.ImageURLs)
		if err != nil {
			return nil, err
		}
	}
	var options []domainproduct.Option
	if req.Options != nil {
		if options, err = productOptions(*req.Options); err != nil {
			return nil, err
		}
	}

	if err := u.tx.WithTx(ctx, func(txCtx context.Context) error {
		_, updateErr := u.products.Update(txCtx, port.UpdateProductParams{
			ID:             id,
			CategoryID:     categoryID,
			Name:           name,
			Slug:           slug,
			Description:    req.Description,
			PriceCents:     req.PriceCents,
			IsActive:       req.IsActive,
			LeadTime:       leadTime,
			IsCustomizable: req.IsCustomizable,
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
		if req.Options != nil {
			if err := u.products.ReplaceOptions(txCtx, id, options); err != nil {
				return err
			}
		}
		if !replaceImages {
			return nil
		}
		return u.products.ReplaceProductImages(txCtx, id, imageURLs)
	}); err != nil {
		return nil, err
	}

	resp, err := u.managerProductResponse(ctx, id)
	if err != nil {
		return nil, err
	}

	u.logAudit(
		ctx,
		domaincatalog.AuditActionManagerUpdatedProduct,
		actorID,
		actorRole,
		&id,
		"product",
		toProductAudit(before, beforeImages, mapProductOptionsToDTO(beforeOptions)),
		toProductAuditFromResponse(resp),
	)
	return resp, nil
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

	u.logAudit(ctx, domaincatalog.AuditActionManagerDeletedProduct, actorID, actorRole, &id, "product", toProductAudit(before, beforeImages, nil), nil)
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
	options, err := u.products.ListOptions(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	resp := toProductResponse(&item.Product, item.CategoryName, imageURLs)
	resp.Options = mapProductOptionsToDTO(options)
	return resp, nil
}

// productOptions reads the options a manager offers, each placed by its order
// within its group.
func productOptions(inputs []dto.ProductOptionInput) ([]domainproduct.Option, error) {
	if len(inputs) > domainproduct.MaxOptions {
		return nil, domainproduct.ErrInvalidOption
	}
	out := make([]domainproduct.Option, 0, len(inputs))
	for _, input := range inputs {
		var id uuid.UUID
		if input.ID != nil {
			parsed, err := uuid.Parse(*input.ID)
			if err != nil {
				return nil, apperrors.ErrValidation.WithDetail("options", "invalid uuid")
			}
			id = parsed
		}
		option, err := domainproduct.NewOption(id, domainproduct.OptionGroup(input.Group), input.Label, input.PriceDeltaCents, input.IsActive)
		if err != nil {
			return nil, err
		}
		out = append(out, option)
	}
	domainproduct.Place(out)
	return out, nil
}

func mapProductOptionsToDTO(options []domainproduct.Option) []dto.ProductOptionResponse {
	out := make([]dto.ProductOptionResponse, 0, len(options))
	for _, option := range options {
		out = append(out, dto.ProductOptionResponse{
			ID:              option.ID.String(),
			Group:           string(option.Group),
			Label:           option.Label,
			PriceDeltaCents: option.PriceDeltaCents,
			IsActive:        option.IsActive,
		})
	}
	return out
}

// productLeadTime reads the notice a product needs, in minutes. The request
// validator bounds it too; the usecase does not rely on that.
func productLeadTime(minutes int32) (time.Duration, error) {
	lead := time.Duration(minutes) * time.Minute
	if lead < 0 || lead > domainproduct.MaxLeadTime {
		return 0, apperrors.ErrValidation.WithDetail("lead_time_minutes", "between 0 and 10080")
	}
	return lead, nil
}

func toProductResponse(product *domainproduct.Product, categoryName string, imageURLs []string) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:              product.ID.String(),
		CategoryID:      product.CategoryID.String(),
		CategoryName:    categoryName,
		Name:            product.Name,
		Slug:            product.Slug,
		Description:     product.Description,
		PriceCents:      product.PriceCents,
		IsActive:        product.IsActive,
		LeadTimeMinutes: utils.Int32FromInt64(int64(product.LeadTime / time.Minute)),
		ImageURLs:       imageURLs,
		IsCustomizable:  product.IsCustomizable,
		CreatedAt:       product.CreatedAt,
		UpdatedAt:       product.UpdatedAt,
	}
}

func toProductAuditFromResponse(resp *dto.ProductResponse) map[string]any {
	return map[string]any{
		"category_id":       resp.CategoryID,
		"name":              resp.Name,
		"slug":              resp.Slug,
		"price_cents":       resp.PriceCents,
		"is_active":         resp.IsActive,
		"lead_time_minutes": resp.LeadTimeMinutes,
		"image_urls":        resp.ImageURLs,
		"is_customizable":   resp.IsCustomizable,
		"options":           resp.Options,
	}
}

func toProductAudit(product *domainproduct.Product, imageURLs []string, options []dto.ProductOptionResponse) map[string]any {
	out := map[string]any{
		"image_urls": imageURLs,
		"options":    options,
	}
	if product != nil {
		out["category_id"] = product.CategoryID.String()
		out["name"] = product.Name
		out["slug"] = product.Slug
		out["price_cents"] = product.PriceCents
		out["is_active"] = product.IsActive
		out["is_customizable"] = product.IsCustomizable
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
