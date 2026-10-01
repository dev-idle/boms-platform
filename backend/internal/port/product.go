package port

import (
	"context"
	"time"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainproduct "github.com/boms/backend/internal/domain/product"
	"github.com/google/uuid"
)

type CreateProductParams struct {
	CategoryID  uuid.UUID
	Name        string
	Slug        string
	Description *string
	PriceCents  int64
	IsActive    bool
	LeadTime    time.Duration
	// IsCustomizable products are configured from their options.
	IsCustomizable bool
}

type UpdateProductParams struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Name        string
	Slug        string
	Description *string
	PriceCents  int64
	IsActive    bool
	LeadTime    time.Duration
	// IsCustomizable products are configured from their options.
	IsCustomizable bool
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
	ImageURLs    []string
}

type CatalogListProduct struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Name        string
	Slug        string
	Description *string
	PriceCents  int64
	// IsCustomizable products are configured from their options before they
	// go in the cart.
	IsCustomizable bool
	CategoryName   string
	CategorySlug   string
	ImageURLs      []string
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
	ReplaceProductImages(ctx context.Context, productID uuid.UUID, imageURLs []string) error
	ListProductImagesByProductID(ctx context.Context, productID uuid.UUID) ([]string, error)
	// FulfillmentOf is what these products and combos ask of the bakery, combos
	// counted by the products they hold.
	FulfillmentOf(ctx context.Context, productIDs, comboIDs []uuid.UUID) (domainorder.Fulfillment, error)
	// ListOptions returns the options these products offer, retired ones left
	// out, by product in the order a customer chooses them.
	ListOptions(ctx context.Context, productIDs []uuid.UUID) ([]domainproduct.Option, error)
	// ReplaceOptions makes options the product's full list: one with an id
	// already the product's is updated, one without is added, and any other is
	// retired. It must run in a transaction; an id the product does not have
	// is ErrInvalidOption.
	ReplaceOptions(ctx context.Context, productID uuid.UUID, options []domainproduct.Option) error
}
