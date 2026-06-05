package port

import (
	"context"

	domaincategory "github.com/boms/backend/internal/domain/category"
	"github.com/google/uuid"
)

type CreateCategoryParams struct {
	Name      string
	Slug      string
	SortOrder int32
	IsActive  bool
}

type UpdateCategoryParams struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	SortOrder int32
	IsActive  bool
}

type ManagerListCategoriesParams struct {
	Search *string
	Limit  int32
	Offset int32
}

type CatalogListCategoriesParams struct {
	Limit  int32
	Offset int32
}

type CategoryRepository interface {
	Create(ctx context.Context, params CreateCategoryParams) (*domaincategory.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domaincategory.Category, error)
	Update(ctx context.Context, params UpdateCategoryParams) (*domaincategory.Category, error)
	SoftDeleteIfNoProducts(ctx context.Context, id uuid.UUID) (int64, error)
	ManagerList(ctx context.Context, params ManagerListCategoriesParams) ([]domaincategory.Category, error)
	ManagerListCount(ctx context.Context, search *string) (int64, error)
	CatalogList(ctx context.Context, params CatalogListCategoriesParams) ([]domaincategory.Category, error)
	CatalogListCount(ctx context.Context) (int64, error)
}
