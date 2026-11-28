-- name: CreateCategory :one
INSERT INTO categories (name, slug, sort_order, is_active, station)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, slug, sort_order, is_active, created_at, updated_at, deleted_at, station;

-- name: GetCategoryByID :one
SELECT id, name, slug, sort_order, is_active, created_at, updated_at, deleted_at, station
FROM categories
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateCategory :one
UPDATE categories
SET name       = $2,
    slug       = $3,
    sort_order = $4,
    is_active  = $5,
    station    = $6,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id, name, slug, sort_order, is_active, created_at, updated_at, deleted_at, station;

-- name: SoftDeleteCategoryIfNoProducts :execrows
UPDATE categories
SET deleted_at = now(),
    updated_at = now()
WHERE categories.id = $1
  AND categories.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1
    FROM products
    WHERE products.category_id = categories.id
      AND products.deleted_at IS NULL
  );

-- name: ManagerListCategories :many
SELECT id, name, slug, sort_order, is_active, created_at, updated_at, deleted_at, station
FROM categories
WHERE deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR slug ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY sort_order ASC, name ASC
LIMIT $1 OFFSET $2;

-- name: ManagerListCategoriesCount :one
SELECT count(*)::bigint AS count
FROM categories
WHERE deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR slug ILIKE '%' || sqlc.narg('search')::text || '%'
  );

-- name: CatalogListCategories :many
-- Each category carries how many products the shop shows in it and a cover:
-- the first photo of the first product (by name) that has one.
SELECT
    c.id,
    c.name,
    c.slug,
    c.sort_order,
    shown.product_count,
    -- sqlc reads a lateral column as non-null; '' stands for "no photo yet".
    COALESCE(cover.image_url, '')::text AS image_url
FROM categories c
CROSS JOIN LATERAL (
    SELECT count(*)::bigint AS product_count
    FROM products p
    WHERE p.category_id = c.id
      AND p.deleted_at IS NULL
      AND p.is_active = true
) shown
LEFT JOIN LATERAL (
    SELECT pi.image_url
    FROM products p
    INNER JOIN product_images pi ON pi.product_id = p.id
    WHERE p.category_id = c.id
      AND p.deleted_at IS NULL
      AND p.is_active = true
    ORDER BY p.name ASC, pi.sort_order ASC
    LIMIT 1
) cover ON true
WHERE c.deleted_at IS NULL
  AND c.is_active = true
ORDER BY c.sort_order ASC, c.name ASC
LIMIT $1 OFFSET $2;

-- name: CatalogListCategoriesCount :one
SELECT count(*)::bigint AS count
FROM categories
WHERE deleted_at IS NULL
  AND is_active = true;
