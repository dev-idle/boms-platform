package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainproduct "github.com/boms/backend/internal/domain/product"
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
	// SoldOutOn is the last bakery day the counter ran out of it; nil if never.
	SoldOutOn    *time.Time
	CategoryName string
	CategorySlug string
	ImageURLs    []string
}

// StaffListProductsParams pages what the counter can mark sold out;
// SoldOutOn narrows the page to the products out on that bakery day.
type StaffListProductsParams struct {
	SoldOutOn *time.Time
	Limit     int32
	Offset    int32
}

// StaffProduct is a product as the counter marks it sold out.
type StaffProduct struct {
	ID           uuid.UUID
	Name         string
	CategoryName string
	Station      domaincategory.Station
	SoldOutOn    *time.Time
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
	FulfillmentOf(ctx context.Context, productIDs, comboIDs []uuid.UUID, today time.Time) (domainorder.Fulfillment, error)
	StaffList(ctx context.Context, params StaffListProductsParams) ([]StaffProduct, error)
	StaffListCount(ctx context.Context, soldOutOn *time.Time) (int64, error)
	// SetSoldOut marks an active product sold out on a bakery day, or back in
	// with nil. Any other product is apperrors.ErrNotFound.
	SetSoldOut(ctx context.Context, id uuid.UUID, day *time.Time) (*StaffProduct, error)
	// ListOptions returns the options these products offer, retired ones left
	// out, by product in the order a customer chooses them.
	ListOptions(ctx context.Context, productIDs []uuid.UUID) ([]domainproduct.Option, error)
	// ReplaceOptions makes options the product's full list: one with an id
	// already the product's is updated, one without is added, and any other is
	// retired. It must run in a transaction; an id the product does not have
	// is ErrInvalidOption.
	ReplaceOptions(ctx context.Context, productID uuid.UUID, options []domainproduct.Option) error
}
