package dto

import "time"

type CreateCategoryRequest struct {
	Name      string `json:"name" validate:"required,max=255"`
	Slug      string `json:"slug" validate:"omitempty,max=128"`
	SortOrder int32  `json:"sort_order" validate:"min=0"`
	IsActive  bool   `json:"is_active"`
}

type UpdateCategoryRequest struct {
	Name      string `json:"name" validate:"required,max=255"`
	Slug      string `json:"slug" validate:"required,max=128"`
	SortOrder int32  `json:"sort_order" validate:"min=0"`
	IsActive  bool   `json:"is_active"`
}

type CategoryResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	SortOrder int32     `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	CategoryID  string  `json:"category_id" validate:"required,uuid"`
	Name        string  `json:"name" validate:"required,max=255"`
	Slug        string  `json:"slug" validate:"omitempty,max=128"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	PriceCents  int64   `json:"price_cents" validate:"min=0"`
	IsAvailable bool    `json:"is_available"`
	ImageURL    *string `json:"image_url,omitempty" validate:"omitempty,max=2048,url"`
}

type UpdateProductRequest struct {
	CategoryID  string  `json:"category_id" validate:"required,uuid"`
	Name        string  `json:"name" validate:"required,max=255"`
	Slug        string  `json:"slug" validate:"required,max=128"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=2000"`
	PriceCents  int64   `json:"price_cents" validate:"min=0"`
	IsAvailable bool    `json:"is_available"`
	ImageURL    *string `json:"image_url,omitempty" validate:"omitempty,max=2048,url"`
}

type ProductResponse struct {
	ID           string    `json:"id"`
	CategoryID   string    `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  *string   `json:"description,omitempty"`
	PriceCents   int64     `json:"price_cents"`
	IsAvailable  bool      `json:"is_available"`
	ImageURL     *string   `json:"image_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CatalogCategoryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	SortOrder int32  `json:"sort_order"`
}

type CatalogProductResponse struct {
	ID           string  `json:"id"`
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	CategorySlug string  `json:"category_slug"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description,omitempty"`
	PriceCents   int64   `json:"price_cents"`
	ImageURL     *string `json:"image_url,omitempty"`
}

type ComboItemInput struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int32  `json:"quantity" validate:"min=1"`
}

type CreateComboRequest struct {
	Name       string           `json:"name" validate:"required,max=255"`
	Slug       string           `json:"slug" validate:"omitempty,max=128"`
	PriceCents int64            `json:"price_cents" validate:"min=0"`
	StartsAt   time.Time        `json:"starts_at" validate:"required"`
	EndsAt     time.Time        `json:"ends_at" validate:"required"`
	IsActive   bool             `json:"is_active"`
	Items      []ComboItemInput `json:"items" validate:"required,min=1,dive"`
}

type UpdateComboRequest struct {
	Name       string           `json:"name" validate:"required,max=255"`
	Slug       string           `json:"slug" validate:"required,max=128"`
	PriceCents int64            `json:"price_cents" validate:"min=0"`
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
	StartsAt   time.Time           `json:"starts_at"`
	EndsAt     time.Time           `json:"ends_at"`
	Items      []ComboItemResponse `json:"items"`
}

type CreateDiscountCodeRequest struct {
	Code          string    `json:"code" validate:"required,max=64"`
	DiscountType  string    `json:"discount_type" validate:"required,oneof=percent fixed_cents"`
	Value         int64     `json:"value" validate:"min=1"`
	MinOrderCents *int64    `json:"min_order_cents,omitempty" validate:"omitempty,min=0"`
	MaxUses       *int32    `json:"max_uses,omitempty" validate:"omitempty,min=1"`
	StartsAt      time.Time `json:"starts_at" validate:"required"`
	EndsAt        time.Time `json:"ends_at" validate:"required"`
	IsActive      bool      `json:"is_active"`
}

type UpdateDiscountCodeRequest struct {
	Code          string    `json:"code" validate:"required,max=64"`
	DiscountType  string    `json:"discount_type" validate:"required,oneof=percent fixed_cents"`
	Value         int64     `json:"value" validate:"min=1"`
	MinOrderCents *int64    `json:"min_order_cents,omitempty" validate:"omitempty,min=0"`
	MaxUses       *int32    `json:"max_uses,omitempty" validate:"omitempty,min=1"`
	StartsAt      time.Time `json:"starts_at" validate:"required"`
	EndsAt        time.Time `json:"ends_at" validate:"required"`
	IsActive      bool      `json:"is_active"`
}

type DiscountCodeResponse struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	DiscountType  string    `json:"discount_type"`
	Value         int64     `json:"value"`
	MinOrderCents *int64    `json:"min_order_cents,omitempty"`
	MaxUses       *int32    `json:"max_uses,omitempty"`
	UsedCount     int32     `json:"used_count"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
