-- name: CreateOrder :one
INSERT INTO orders (
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  created_at,
  updated_at;

-- name: GetOrderByIDForUser :one
SELECT
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  created_at,
  updated_at
FROM orders
WHERE id = $1 AND user_id = $2;

-- name: ListOrdersByUser :many
SELECT
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  created_at,
  updated_at
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrdersByUserCount :one
SELECT COUNT(*)::bigint AS count
FROM orders
WHERE user_id = $1;

-- name: CreateOrderItem :one
INSERT INTO order_items (
  order_id,
  line_type,
  product_id,
  combo_id,
  name,
  slug,
  quantity,
  unit_price_cents,
  line_total_cents
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING
  id,
  order_id,
  line_type,
  product_id,
  combo_id,
  name,
  slug,
  quantity,
  unit_price_cents,
  line_total_cents,
  created_at;

-- name: SumOrderItemQuantitiesByOrderIDs :many
SELECT order_id, COALESCE(SUM(quantity), 0)::bigint AS item_count
FROM order_items
WHERE order_id = ANY(sqlc.arg('order_ids')::uuid[])
GROUP BY order_id;

-- name: ListOrderItemsByOrderID :many
SELECT
  id,
  order_id,
  line_type,
  product_id,
  combo_id,
  name,
  slug,
  quantity,
  unit_price_cents,
  line_total_cents,
  created_at
FROM order_items
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: StaffListOrders :many
SELECT
  o.id,
  o.user_id,
  o.status,
  o.subtotal_cents,
  o.discount_cents,
  o.total_cents,
  o.discount_code_id,
  o.discount_code_snapshot,
  o.created_at,
  o.updated_at,
  u.email AS customer_email,
  cp.display_name AS customer_display_name
FROM orders o
INNER JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
LEFT JOIN customer_profiles cp ON cp.user_id = o.user_id
WHERE (
    sqlc.narg('status')::order_status IS NULL
    OR o.status = sqlc.narg('status')::order_status
)
ORDER BY o.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: StaffListOrdersCount :one
SELECT COUNT(*)::bigint AS count
FROM orders o
INNER JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
WHERE (
    sqlc.narg('status')::order_status IS NULL
    OR o.status = sqlc.narg('status')::order_status
);

-- name: StaffGetOrderByID :one
SELECT
  o.id,
  o.user_id,
  o.status,
  o.subtotal_cents,
  o.discount_cents,
  o.total_cents,
  o.discount_code_id,
  o.discount_code_snapshot,
  o.created_at,
  o.updated_at,
  u.email AS customer_email,
  cp.display_name AS customer_display_name
FROM orders o
INNER JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
LEFT JOIN customer_profiles cp ON cp.user_id = o.user_id
WHERE o.id = $1;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = sqlc.arg('to_status')::order_status,
    updated_at = now()
WHERE id = sqlc.arg('id')
  AND status = sqlc.arg('from_status')::order_status
RETURNING
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  created_at,
  updated_at;
