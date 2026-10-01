package dto

import "time"

// SavedProductResponse is one product on a customer's favorites or wishlist,
// as the catalog shows it.
type SavedProductResponse struct {
	List    string                 `json:"list"`
	SavedAt time.Time              `json:"saved_at"`
	Product CatalogProductResponse `json:"product"`
}
