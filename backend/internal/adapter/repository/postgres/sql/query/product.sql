-- name: CreateProduct :one
INSERT INTO products (category_id, name, slug, description, price_cents, is_available, image_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, category_id, name, slug, description, price_cents, is_available, image_url, created_at, updated_at, deleted_at;

-- name: GetProductByID :one
SELECT id, category_id, name, slug, description, price_cents, is_available, image_url, created_at, updated_at, deleted_at
FROM products
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ManagerGetProductByID :one
SELECT
    p.id,
    p.category_id,
    p.name,
    p.slug,
    p.description,
    p.price_cents,
    p.is_available,
    p.image_url,
    p.created_at,
    p.updated_at,
    p.deleted_at,
    c.name AS category_name
FROM products p
INNER JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
WHERE p.id = $1
  AND p.deleted_at IS NULL;

-- name: UpdateProduct :one
UPDATE products
SET category_id  = $2,
    name         = $3,
    slug         = $4,
    description  = $5,
    price_cents  = $6,
    is_available = $7,
    image_url    = $8,
    updated_at   = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id, category_id, name, slug, description, price_cents, is_available, image_url, created_at, updated_at, deleted_at;

-- name: SoftDeleteProduct :execrows
UPDATE products
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ManagerListProducts :many
SELECT
    p.id,
    p.category_id,
    p.name,
    p.slug,
    p.description,
    p.price_cents,
    p.is_available,
    p.image_url,
    p.created_at,
    p.updated_at,
    p.deleted_at,
    c.name AS category_name
FROM products p
INNER JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL
WHERE p.deleted_at IS NULL
  AND (
    sqlc.narg('category_id')::uuid IS NULL
    OR p.category_id = sqlc.narg('category_id')::uuid
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR p.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR p.slug ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY p.name ASC
LIMIT $1 OFFSET $2;

-- name: ManagerListProductsCount :one
SELECT count(*)::bigint AS count
FROM products p
WHERE p.deleted_at IS NULL
  AND (
    sqlc.narg('category_id')::uuid IS NULL
    OR p.category_id = sqlc.narg('category_id')::uuid
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR p.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR p.slug ILIKE '%' || sqlc.narg('search')::text || '%'
  );

-- name: CatalogListProducts :many
SELECT
    p.id,
    p.category_id,
    p.name,
    p.slug,
    p.description,
    p.price_cents,
    p.image_url,
    c.name AS category_name,
    c.slug AS category_slug
FROM products p
INNER JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL AND c.is_active = true
WHERE p.deleted_at IS NULL
  AND p.is_available = true
  AND (
    sqlc.narg('category_id')::uuid IS NULL
    OR p.category_id = sqlc.narg('category_id')::uuid
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR p.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR p.slug ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY c.sort_order ASC, p.name ASC
LIMIT $1 OFFSET $2;

-- name: CatalogListProductsCount :one
SELECT count(*)::bigint AS count
FROM products p
INNER JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL AND c.is_active = true
WHERE p.deleted_at IS NULL
  AND p.is_available = true
  AND (
    sqlc.narg('category_id')::uuid IS NULL
    OR p.category_id = sqlc.narg('category_id')::uuid
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR p.name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR p.slug ILIKE '%' || sqlc.narg('search')::text || '%'
  );

-- name: CatalogGetProductsByIDs :many
SELECT
    p.id,
    p.category_id,
    p.name,
    p.slug,
    p.description,
    p.price_cents,
    p.image_url,
    c.name AS category_name,
    c.slug AS category_slug
FROM products p
INNER JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL AND c.is_active = true
WHERE p.id = ANY(sqlc.arg('product_ids')::uuid[])
  AND p.deleted_at IS NULL
  AND p.is_available = true;

-- name: CatalogGetProductByID :one
SELECT
    p.id,
    p.category_id,
    p.name,
    p.slug,
    p.description,
    p.price_cents,
    p.image_url,
    c.name AS category_name,
    c.slug AS category_slug
FROM products p
INNER JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL AND c.is_active = true
WHERE p.id = $1
  AND p.deleted_at IS NULL
  AND p.is_available = true;
