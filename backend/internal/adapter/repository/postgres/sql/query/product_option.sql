-- name: ListProductOptions :many
-- The options these products offer, retired ones left out, in the order a
-- customer chooses them.
SELECT id, product_id, option_group, label, price_delta_cents, sort_order, is_active
FROM product_options
WHERE product_id = ANY(sqlc.arg('product_ids')::uuid[])
  AND deleted_at IS NULL
ORDER BY product_id, option_group, sort_order;

-- name: RetireProductOptions :exec
-- Retires the product's options the manager left out. Cart lines may still
-- name them, so they stay as rows.
UPDATE product_options
SET deleted_at = now(),
    updated_at = now()
WHERE product_id = sqlc.arg('product_id')
  AND deleted_at IS NULL
  AND NOT (id = ANY(sqlc.arg('kept_ids')::uuid[]));

-- name: UpdateProductOptions :execrows
-- The manager's edits to options the product already has, in one statement.
UPDATE product_options o
SET option_group = r.option_group,
    label = r.label,
    price_delta_cents = r.price_delta_cents,
    sort_order = r.sort_order,
    is_active = r.is_active,
    updated_at = now()
FROM jsonb_to_recordset(sqlc.arg('rows')::jsonb) AS r (
  id uuid,
  option_group product_option_group,
  label text,
  price_delta_cents bigint,
  sort_order smallint,
  is_active boolean
)
WHERE o.id = r.id
  AND o.product_id = sqlc.arg('product_id')
  AND o.deleted_at IS NULL;

-- name: CreateProductOptions :execrows
-- The options the manager added, in one statement.
INSERT INTO product_options (product_id, option_group, label, price_delta_cents, sort_order, is_active)
SELECT sqlc.arg('product_id'), r.option_group, r.label, r.price_delta_cents, r.sort_order, r.is_active
FROM jsonb_to_recordset(sqlc.arg('rows')::jsonb) AS r (
  option_group product_option_group,
  label text,
  price_delta_cents bigint,
  sort_order smallint,
  is_active boolean
);
