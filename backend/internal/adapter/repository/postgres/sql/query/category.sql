-- name: CreateCategory :one
INSERT INTO categories (name, slug, sort_order, is_active)
VALUES ($1, $2, $3, $4)
RETURNING id, name, slug, sort_order, is_active, created_at, updated_at, deleted_at;

-- name: GetCategoryByID :one
SELECT id, name, slug, sort_order, is_active, created_at, updated_at, deleted_at
FROM categories
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateCategory :one
UPDATE categories
SET name       = $2,
    slug       = $3,
    sort_order = $4,
    is_active  = $5,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id, name, slug, sort_order, is_active, created_at, updated_at, deleted_at;

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
SELECT id, name, slug, sort_order, is_active, created_at, updated_at, deleted_at
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
SELECT id, name, slug, sort_order
FROM categories
WHERE deleted_at IS NULL
  AND is_active = true
ORDER BY sort_order ASC, name ASC
LIMIT $1 OFFSET $2;

-- name: CatalogListCategoriesCount :one
SELECT count(*)::bigint AS count
FROM categories
WHERE deleted_at IS NULL
  AND is_active = true;
