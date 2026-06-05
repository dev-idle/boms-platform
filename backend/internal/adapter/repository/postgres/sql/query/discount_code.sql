-- name: CreateDiscountCode :one
INSERT INTO discount_codes (
    code,
    discount_type,
    value,
    min_order_cents,
    max_uses,
    starts_at,
    ends_at,
    is_active
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
    id,
    code,
    discount_type,
    value,
    min_order_cents,
    max_uses,
    used_count,
    starts_at,
    ends_at,
    is_active,
    created_at,
    updated_at,
    deleted_at;

-- name: GetDiscountCodeByID :one
SELECT
    id,
    code,
    discount_type,
    value,
    min_order_cents,
    max_uses,
    used_count,
    starts_at,
    ends_at,
    is_active,
    created_at,
    updated_at,
    deleted_at
FROM discount_codes
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateDiscountCode :one
UPDATE discount_codes
SET code            = $2,
    discount_type   = $3,
    value           = $4,
    min_order_cents = $5,
    max_uses        = $6,
    starts_at       = $7,
    ends_at         = $8,
    is_active       = $9,
    updated_at      = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING
    id,
    code,
    discount_type,
    value,
    min_order_cents,
    max_uses,
    used_count,
    starts_at,
    ends_at,
    is_active,
    created_at,
    updated_at,
    deleted_at;

-- name: SoftDeleteDiscountCode :execrows
UPDATE discount_codes
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ManagerListDiscountCodes :many
SELECT
    id,
    code,
    discount_type,
    value,
    min_order_cents,
    max_uses,
    used_count,
    starts_at,
    ends_at,
    is_active,
    created_at,
    updated_at,
    deleted_at
FROM discount_codes
WHERE deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR code ILIKE '%' || sqlc.narg('search')::text || '%'
  )
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ManagerListDiscountCodesCount :one
SELECT COUNT(*)::bigint AS count
FROM discount_codes
WHERE deleted_at IS NULL
  AND (
    sqlc.narg('search')::text IS NULL
    OR code ILIKE '%' || sqlc.narg('search')::text || '%'
  );
