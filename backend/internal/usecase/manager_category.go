package usecase

import (
	"context"
	"errors"
	"strings"

	domaincatalog "github.com/boms/backend/internal/domain/catalog"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ManagerCategoryUsecase struct {
	categories port.CategoryRepository
	audit      *auditlogger.Service
	log        *zap.Logger
}

func NewManagerCategoryUsecase(categories port.CategoryRepository, audit *auditlogger.Service, log *zap.Logger) *ManagerCategoryUsecase {
	return &ManagerCategoryUsecase{categories: categories, audit: audit, log: log}
}

func (u *ManagerCategoryUsecase) Create(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	req dto.CreateCategoryRequest,
) (*dto.CategoryResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperrors.ErrValidation.WithDetail("name", "required")
	}
	slug, err := resolveManagerCatalogSlug(name, req.Slug, true)
	if err != nil {
		return nil, err
	}

	created, err := u.categories.Create(ctx, port.CreateCategoryParams{
		Name:      name,
		Slug:      slug,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domaincategory.ErrSlugExists
		}
		return nil, err
	}

	u.logAudit(ctx, domaincatalog.AuditActionManagerCreatedCategory, actorID, actorRole, &created.ID, "category", nil, toCategoryAudit(created))
	return toCategoryResponse(created), nil
}

func (u *ManagerCategoryUsecase) Get(ctx context.Context, id uuid.UUID) (*dto.CategoryResponse, error) {
	cat, err := u.categories.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaincategory.ErrNotFound
		}
		return nil, err
	}
	return toCategoryResponse(cat), nil
}

func (u *ManagerCategoryUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	search string,
) ([]dto.CategoryResponse, int64, int32, int32, error) {
	page, pageSize = normalizeCatalogListPage(page, pageSize)
	var searchPtr *string
	trimmed := strings.TrimSpace(search)
	if trimmed != "" {
		searchPtr = &trimmed
	}
	items, err := u.categories.ManagerList(ctx, port.ManagerListCategoriesParams{
		Search: searchPtr,
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.categories.ManagerListCount(ctx, searchPtr)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	out := make([]dto.CategoryResponse, 0, len(items))
	for i := range items {
		out = append(out, *toCategoryResponse(&items[i]))
	}
	return out, total, page, pageSize, nil
}

func (u *ManagerCategoryUsecase) Update(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
	req dto.UpdateCategoryRequest,
) (*dto.CategoryResponse, error) {
	before, err := u.categories.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaincategory.ErrNotFound
		}
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

	updated, err := u.categories.Update(ctx, port.UpdateCategoryParams{
		ID:        id,
		Name:      name,
		Slug:      slug,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, domaincategory.ErrSlugExists
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domaincategory.ErrNotFound
		}
		return nil, err
	}

	u.logAudit(ctx, domaincatalog.AuditActionManagerUpdatedCategory, actorID, actorRole, &id, "category", toCategoryAudit(before), toCategoryAudit(updated))
	return toCategoryResponse(updated), nil
}

func (u *ManagerCategoryUsecase) Delete(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	id uuid.UUID,
) error {
	before, err := u.categories.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return domaincategory.ErrNotFound
		}
		return err
	}

	rows, err := u.categories.SoftDeleteIfNoProducts(ctx, id)
	if err != nil {
		return err
	}
	if rows == 0 {
		if _, getErr := u.categories.GetByID(ctx, id); getErr != nil {
			if errors.Is(getErr, apperrors.ErrNotFound) {
				return domaincategory.ErrNotFound
			}
			return getErr
		}
		return domaincategory.ErrHasProducts
	}

	u.logAudit(ctx, domaincatalog.AuditActionManagerDeletedCategory, actorID, actorRole, &id, "category", toCategoryAudit(before), nil)
	return nil
}

func toCategoryResponse(cat *domaincategory.Category) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:        cat.ID.String(),
		Name:      cat.Name,
		Slug:      cat.Slug,
		SortOrder: cat.SortOrder,
		IsActive:  cat.IsActive,
		CreatedAt: cat.CreatedAt,
		UpdatedAt: cat.UpdatedAt,
	}
}

func toCategoryAudit(cat *domaincategory.Category) map[string]any {
	return map[string]any{
		"name":       cat.Name,
		"slug":       cat.Slug,
		"sort_order": cat.SortOrder,
		"is_active":  cat.IsActive,
	}
}

func (u *ManagerCategoryUsecase) logAudit(
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
