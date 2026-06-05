package dto

import "time"

type CreateCategoryRequest struct {
	Name      string `json:"name" validate:"required,max=255"`
	Slug      string `json:"slug" validate:"required,max=128"`
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
	Slug        string  `json:"slug" validate:"required,max=128"`
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
