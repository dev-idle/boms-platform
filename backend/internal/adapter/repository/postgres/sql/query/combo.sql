-- name: CreateCombo :one
INSERT INTO combos (name, slug, price_cents, starts_at, ends_at, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, slug, price_cents, starts_at, ends_at, is_active, created_at, updated_at, deleted_at;

-- name: GetComboByID :one
SELECT id, name, slug, price_cents, starts_at, ends_at, is_active, created_at, updated_at, deleted_at
FROM combos
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateCombo :one
UPDATE combos
SET name        = $2,
    slug        = $3,
    price_cents = $4,
    starts_at   = $5,
    ends_at     = $6,
    is_active   = $7,
    updated_at  = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id, name, slug, price_cents, starts_at, ends_at, is_active, created_at, updated_at, deleted_at;

-- name: SoftDeleteCombo :execrows
UPDATE combos
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: DeleteComboItemsByComboID :exec
DELETE FROM combo_items
WHERE combo_id = $1;

-- name: InsertComboItem :exec
INSERT INTO combo_items (combo_id, product_id, quantity)
VALUES ($1, $2, $3);

-- name: ListComboItemsByComboID :many
SELECT
    ci.id,
    ci.combo_id,
    ci.product_id,
    ci.quantity,
    p.name AS product_name,
    p.slug AS product_slug,
    p.price_cents
FROM combo_items ci
INNER JOIN products p ON p.id = ci.product_id AND p.deleted_at IS NULL
WHERE ci.combo_id = $1
ORDER BY p.name ASC;

-- name: ListComboItemsByComboIDs :many
SELECT
    ci.id,
    ci.combo_id,
    ci.product_id,
    ci.quantity,
    p.name AS product_name,
    p.slug AS product_slug,
    p.price_cents
FROM combo_items ci
INNER JOIN products p ON p.id = ci.product_id AND p.deleted_at IS NULL
WHERE ci.combo_id = ANY(sqlc.arg('combo_ids')::uuid[])
ORDER BY ci.combo_id ASC, p.name ASC;

-- name: ListCatalogComboItemsByComboIDs :many
SELECT
    ci.id,
    ci.combo_id,
    ci.product_id,
    ci.quantity,
    p.name AS product_name,
    p.slug AS product_slug,
    p.price_cents
FROM combo_items ci
INNER JOIN products p ON p.id = ci.product_id AND p.deleted_at IS NULL AND p.is_available = true
INNER JOIN categories cat ON cat.id = p.category_id AND cat.deleted_at IS NULL AND cat.is_active = true
WHERE ci.combo_id = ANY(sqlc.arg('combo_ids')::uuid[])
ORDER BY ci.combo_id ASC, p.name ASC;

-- name: ManagerListCombos :many
SELECT id, name, slug, price_cents, starts_at, ends_at, is_active, created_at, updated_at, deleted_at
FROM combos
WHERE deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR slug ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY starts_at DESC, name ASC
LIMIT $1 OFFSET $2;

-- name: ManagerListCombosCount :one
SELECT COUNT(*)::bigint AS count
FROM combos
WHERE deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR name ILIKE '%' || sqlc.narg('search')::text || '%'
    OR slug ILIKE '%' || sqlc.narg('search')::text || '%'
  );

-- name: CatalogListCombos :many
SELECT c.id, c.name, c.slug, c.price_cents, c.starts_at, c.ends_at
FROM combos c
WHERE c.deleted_at IS NULL
  AND c.is_active = true
  AND c.starts_at <= now()
  AND c.ends_at > now()
  AND EXISTS (
    SELECT 1
    FROM combo_items ci
    INNER JOIN products p ON p.id = ci.product_id AND p.deleted_at IS NULL AND p.is_available = true
    INNER JOIN categories cat ON cat.id = p.category_id AND cat.deleted_at IS NULL AND cat.is_active = true
    WHERE ci.combo_id = c.id
  )
ORDER BY c.starts_at ASC, c.name ASC
LIMIT $1 OFFSET $2;

-- name: CatalogListCombosCount :one
SELECT COUNT(*)::bigint AS count
FROM combos c
WHERE c.deleted_at IS NULL
  AND c.is_active = true
  AND c.starts_at <= now()
  AND c.ends_at > now()
  AND EXISTS (
    SELECT 1
    FROM combo_items ci
    INNER JOIN products p ON p.id = ci.product_id AND p.deleted_at IS NULL AND p.is_available = true
    INNER JOIN categories cat ON cat.id = p.category_id AND cat.deleted_at IS NULL AND cat.is_active = true
    WHERE ci.combo_id = c.id
  );

-- name: CatalogGetCombosByIDs :many
SELECT c.id, c.name, c.slug, c.price_cents, c.starts_at, c.ends_at
FROM combos c
WHERE c.id = ANY(sqlc.arg('combo_ids')::uuid[])
  AND c.deleted_at IS NULL
  AND c.is_active = true
  AND c.starts_at <= now()
  AND c.ends_at > now()
  AND EXISTS (
    SELECT 1
    FROM combo_items ci
    INNER JOIN products p ON p.id = ci.product_id AND p.deleted_at IS NULL AND p.is_available = true
    INNER JOIN categories cat ON cat.id = p.category_id AND cat.deleted_at IS NULL AND cat.is_active = true
    WHERE ci.combo_id = c.id
  );

-- name: CatalogGetComboByID :one
SELECT id, name, slug, price_cents, starts_at, ends_at
FROM combos
WHERE id = $1
  AND deleted_at IS NULL
  AND is_active = true
  AND starts_at <= now()
  AND ends_at > now();

-- name: CountAvailableProductsForCombo :one
SELECT COUNT(DISTINCT p.id)::bigint AS count
FROM products p
INNER JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL AND c.is_active = true
WHERE p.id = ANY(sqlc.arg('product_ids')::uuid[])
  AND p.deleted_at IS NULL
  AND p.is_available = true;
