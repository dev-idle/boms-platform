package dto

import "time"

type CreateCategoryRequest struct {
	Name      string `json:"name" validate:"required,max=255"`
	Slug      string `json:"slug" validate:"omitempty,max=128"`
	SortOrder int32  `json:"sort_order" validate:"min=0"`
	IsActive  bool   `json:"is_active"`
	Station   string `json:"station" validate:"required,oneof=kitchen counter"`
}

type UpdateCategoryRequest struct {
	Name      string `json:"name" validate:"required,max=255"`
	Slug      string `json:"slug" validate:"required,max=128"`
	SortOrder int32  `json:"sort_order" validate:"min=0"`
	IsActive  bool   `json:"is_active"`
	Station   string `json:"station" validate:"required,oneof=kitchen counter"`
}

type CategoryResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	SortOrder int32     `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	Station   string    `json:"station"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	CategoryID  string  `json:"category_id" validate:"required,uuid"`
	Name        string  `json:"name" validate:"required,max=255"`
	Slug        string  `json:"slug" validate:"omitempty,max=128"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	PriceCents  int64   `json:"price_cents" validate:"min=0"`
	IsActive    bool    `json:"is_active"`
	// LeadTimeMinutes is the notice the product needs before pickup, up to a week.
	LeadTimeMinutes int32    `json:"lead_time_minutes" validate:"min=0,max=10080"`
	ImageURLs       []string `json:"image_urls,omitempty" validate:"omitempty,max=5,dive,url,max=2048"`
	IsCustomizable  bool     `json:"is_customizable"`
	// Options are what a customer chooses from on a customizable product.
	Options []ProductOptionInput `json:"options,omitempty" validate:"omitempty,max=30,dive"`
}

// ProductOptionInput is one option a manager offers on a customizable product.
// Its place within its group is its place in the list; one without an id is
// new, and an option left out of the list is retired.
type ProductOptionInput struct {
	ID              *string `json:"id,omitempty" validate:"omitempty,uuid"`
	Group           string  `json:"group" validate:"required,oneof=size flavor decoration"`
	Label           string  `json:"label" validate:"required,max=60"`
	PriceDeltaCents int64   `json:"price_delta_cents" validate:"min=0,max=100000"`
	IsActive        bool    `json:"is_active"`
}

type UpdateProductRequest struct {
	CategoryID  string  `json:"category_id" validate:"required,uuid"`
	Name        string  `json:"name" validate:"required,max=255"`
	Slug        string  `json:"slug" validate:"required,max=128"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	PriceCents  int64   `json:"price_cents" validate:"min=0"`
	IsActive    bool    `json:"is_active"`
	// LeadTimeMinutes is the notice the product needs before pickup, up to a week.
	LeadTimeMinutes int32     `json:"lead_time_minutes" validate:"min=0,max=10080"`
	ImageURLs       *[]string `json:"image_urls,omitempty" validate:"omitempty,max=5,dive,url,max=2048"`
	IsCustomizable  bool      `json:"is_customizable"`
	// Options, when sent, become the product's full list of options.
	Options *[]ProductOptionInput `json:"options,omitempty" validate:"omitempty,max=30,dive"`
}

type ProductResponse struct {
	ID              string   `json:"id"`
	CategoryID      string   `json:"category_id"`
	CategoryName    string   `json:"category_name,omitempty"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	Description     *string  `json:"description,omitempty"`
	PriceCents      int64    `json:"price_cents"`
	IsActive        bool     `json:"is_active"`
	LeadTimeMinutes int32    `json:"lead_time_minutes"`
	ImageURLs       []string `json:"image_urls,omitempty"`
	IsCustomizable  bool     `json:"is_customizable"`
	// Options are the product's options, retired ones left out; on its detail only.
	Options   []ProductOptionResponse `json:"options,omitempty"`
	CreatedAt time.Time               `json:"created_at"`
	UpdatedAt time.Time               `json:"updated_at"`
}

type ProductOptionResponse struct {
	ID              string `json:"id"`
	Group           string `json:"group"`
	Label           string `json:"label"`
	PriceDeltaCents int64  `json:"price_delta_cents"`
	IsActive        bool   `json:"is_active"`
}

type CatalogCategoryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	SortOrder int32  `json:"sort_order"`
}

type CatalogProductResponse struct {
	ID           string   `json:"id"`
	CategoryID   string   `json:"category_id"`
	CategoryName string   `json:"category_name"`
	CategorySlug string   `json:"category_slug"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	Description  *string  `json:"description,omitempty"`
	PriceCents   int64    `json:"price_cents"`
	ImageURLs    []string `json:"image_urls,omitempty"`
	// IsCustomizable products are configured before they go in the cart.
	IsCustomizable bool `json:"is_customizable"`
	// SoldOutToday: the counter ran out of it; it may be ordered for a later day.
	SoldOutToday bool `json:"sold_out_today"`
	// Options are what a customer chooses from, on a customizable product's detail.
	Options []CatalogProductOptionResponse `json:"options,omitempty"`
}

// CatalogProductOptionResponse is one option a customer may choose, with what
// it adds to the price.
type CatalogProductOptionResponse struct {
	ID              string `json:"id"`
	Group           string `json:"group"`
	Label           string `json:"label"`
	PriceDeltaCents int64  `json:"price_delta_cents"`
}

type ComboItemInput struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int32  `json:"quantity" validate:"min=1"`
}

type CreateComboRequest struct {
	Name       string           `json:"name" validate:"required,max=255"`
	Slug       string           `json:"slug" validate:"omitempty,max=128"`
	PriceCents int64            `json:"price_cents" validate:"min=0"`
	ImageURL   *string          `json:"image_url,omitempty" validate:"omitempty,url,max=2048"`
	StartsAt   time.Time        `json:"starts_at" validate:"required"`
	EndsAt     time.Time        `json:"ends_at" validate:"required"`
	IsActive   bool             `json:"is_active"`
	Items      []ComboItemInput `json:"items" validate:"required,min=1,dive"`
}

type UpdateComboRequest struct {
	Name       string           `json:"name" validate:"required,max=255"`
	Slug       string           `json:"slug" validate:"required,max=128"`
	PriceCents int64            `json:"price_cents" validate:"min=0"`
	ImageURL   *string          `json:"image_url,omitempty" validate:"omitempty,url,max=2048"`
	StartsAt   time.Time        `json:"starts_at" validate:"required"`
	EndsAt     time.Time        `json:"ends_at" validate:"required"`
	IsActive   bool             `json:"is_active"`
	Items      []ComboItemInput `json:"items" validate:"required,min=1,dive"`
}

type ComboItemResponse struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	ProductSlug string `json:"product_slug"`
	Quantity    int32  `json:"quantity"`
	PriceCents  int64  `json:"price_cents"`
}

type ComboResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Slug       string              `json:"slug"`
	PriceCents int64               `json:"price_cents"`
	ImageURL   *string             `json:"image_url,omitempty"`
	StartsAt   time.Time           `json:"starts_at"`
	EndsAt     time.Time           `json:"ends_at"`
	IsActive   bool                `json:"is_active"`
	Items      []ComboItemResponse `json:"items"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

type CatalogComboResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	Slug       string              `json:"slug"`
	PriceCents int64               `json:"price_cents"`
	ImageURL   *string             `json:"image_url,omitempty"`
	StartsAt   time.Time           `json:"starts_at"`
	EndsAt     time.Time           `json:"ends_at"`
	Items      []ComboItemResponse `json:"items"`
	// SoldOutToday: the counter ran out of one of its products today.
	SoldOutToday bool `json:"sold_out_today"`
}

type CreateDiscountCodeRequest struct {
	Code          string `json:"code" validate:"required,max=64"`
	DiscountType  string `json:"discount_type" validate:"required,oneof=percent fixed_cents"`
	Value         int64  `json:"value" validate:"min=1"`
	MinOrderCents *int64 `json:"min_order_cents,omitempty" validate:"omitempty,min=0"`
	MaxUses       *int32 `json:"max_uses,omitempty"`
	// MaxUsesPerCustomer caps how many of one customer's orders may use the code.
	MaxUsesPerCustomer *int32    `json:"max_uses_per_customer,omitempty" validate:"omitempty,min=1"`
	MaxDiscountCents   *int64    `json:"max_discount_cents,omitempty" validate:"omitempty,min=1"`
	StartsAt           time.Time `json:"starts_at" validate:"required"`
	EndsAt             time.Time `json:"ends_at" validate:"required"`
	IsActive           bool      `json:"is_active"`
}

type UpdateDiscountCodeRequest struct {
	Code          string `json:"code" validate:"required,max=64"`
	DiscountType  string `json:"discount_type" validate:"required,oneof=percent fixed_cents"`
	Value         int64  `json:"value" validate:"min=1"`
	MinOrderCents *int64 `json:"min_order_cents,omitempty" validate:"omitempty,min=0"`
	MaxUses       *int32 `json:"max_uses,omitempty"`
	// MaxUsesPerCustomer caps how many of one customer's orders may use the code.
	MaxUsesPerCustomer *int32    `json:"max_uses_per_customer,omitempty" validate:"omitempty,min=1"`
	MaxDiscountCents   *int64    `json:"max_discount_cents,omitempty" validate:"omitempty,min=1"`
	StartsAt           time.Time `json:"starts_at" validate:"required"`
	EndsAt             time.Time `json:"ends_at" validate:"required"`
	IsActive           bool      `json:"is_active"`
}

type DiscountCodeResponse struct {
	ID                 string    `json:"id"`
	Code               string    `json:"code"`
	DiscountType       string    `json:"discount_type"`
	Value              int64     `json:"value"`
	MinOrderCents      *int64    `json:"min_order_cents,omitempty"`
	MaxUses            *int32    `json:"max_uses,omitempty"`
	MaxUsesPerCustomer *int32    `json:"max_uses_per_customer,omitempty"`
	MaxDiscountCents   *int64    `json:"max_discount_cents,omitempty"`
	UsedCount          int32     `json:"used_count"`
	StartsAt           time.Time `json:"starts_at"`
	EndsAt             time.Time `json:"ends_at"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
