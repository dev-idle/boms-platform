package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainsaved "github.com/boms/backend/internal/domain/saved"
)

// SavedProductRepository keeps the products customers keep on their lists.
type SavedProductRepository interface {
	// Save puts a product on sale on the list, or keeps it there; it reports
	// false for a product not on sale.
	Save(ctx context.Context, userID, productID uuid.UUID, list domainsaved.List) (bool, error)
	// CountOthers counts the products on sale on the list besides productID.
	CountOthers(ctx context.Context, userID, productID uuid.UUID, list domainsaved.List) (int64, error)
	Remove(ctx context.Context, userID, productID uuid.UUID, list domainsaved.List) error
	// ListByUser returns the products on sale on both of the customer's
	// lists, latest first.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]SavedProduct, error)
	// ListForExport returns every product on the customer's lists, on sale or not.
	ListForExport(ctx context.Context, userID uuid.UUID) ([]SavedProductEntry, error)
	// RemoveAll empties the customer's lists.
	RemoveAll(ctx context.Context, userID uuid.UUID) error
}

// SavedProduct is one product on a customer's list, as the catalog shows it.
type SavedProduct struct {
	List    domainsaved.List
	SavedAt time.Time
	Product CatalogListProduct
}

// SavedProductEntry is what a customer's list holds about one product.
type SavedProductEntry struct {
	List        domainsaved.List
	SavedAt     time.Time
	ProductID   uuid.UUID
	ProductName string
}
