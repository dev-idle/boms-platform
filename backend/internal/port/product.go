package port

import (
	"context"

	domainproduct "github.com/boms/backend/internal/domain/product"
	"github.com/google/uuid"
)

type CreateProductParams struct {
	CategoryID  uuid.UUID
	Name        string
	Slug        string
	Description *string
	PriceCents  int64
	IsAvailable bool
	ImageURL    *string
}

type UpdateProductParams struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Name        string
	Slug        string
	Description *string
	PriceCents  int64
	IsAvailable bool
	ImageURL    *string
}

type ManagerListProductsParams struct {
	CategoryID *uuid.UUID
	Search     *string
	Limit      int32
	Offset     int32
}

type CatalogListProductsParams struct {
	CategoryID *uuid.UUID
	Search     *string
	Limit      int32
	Offset     int32
}

type ManagerListProduct struct {
	Product      domainproduct.Product
	CategoryName string
}

type CatalogListProduct struct {
	ID           uuid.UUID
	CategoryID   uuid.UUID
	Name         string
	Slug         string
	Description  *string
	PriceCents   int64
	ImageURL     *string
	CategoryName string
	CategorySlug string
}

type ProductRepository interface {
	Create(ctx context.Context, params CreateProductParams) (*domainproduct.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domainproduct.Product, error)
	Update(ctx context.Context, params UpdateProductParams) (*domainproduct.Product, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ManagerList(ctx context.Context, params ManagerListProductsParams) ([]ManagerListProduct, error)
	ManagerListCount(ctx context.Context, categoryID *uuid.UUID, search *string) (int64, error)
	ManagerGetByID(ctx context.Context, id uuid.UUID) (*ManagerListProduct, error)
	CatalogList(ctx context.Context, params CatalogListProductsParams) ([]CatalogListProduct, error)
	CatalogListCount(ctx context.Context, categoryID *uuid.UUID, search *string) (int64, error)
	CatalogGetByID(ctx context.Context, id uuid.UUID) (*CatalogListProduct, error)
	CatalogGetByIDs(ctx context.Context, ids []uuid.UUID) ([]CatalogListProduct, error)
}
