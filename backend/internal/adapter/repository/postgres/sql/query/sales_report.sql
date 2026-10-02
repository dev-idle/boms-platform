-- name: SalesByPeriod :many
-- The payments taken and the refunds made in [from_at, to_at), by bakery day,
-- week (from Monday) or month, every period of [from_day, to_day] listed even
-- when nothing happened; the first period starts no earlier than from_day.
WITH periods AS (
  SELECT generate_series(
    date_trunc(sqlc.arg('grouping')::text, sqlc.arg('from_day')::date::timestamp),
    sqlc.arg('to_day')::date::timestamp,
    ('1 ' || sqlc.arg('grouping')::text)::interval
  )::date AS start
),
taken AS (
  SELECT
    date_trunc(sqlc.arg('grouping')::text, pay.captured_at AT TIME ZONE sqlc.arg('zone')::text)::date AS start,
    count(*) AS orders,
    sum(pay.amount_cents) AS gross_cents
  FROM payments pay
  WHERE pay.captured_at >= sqlc.arg('from_at')::timestamptz
    AND pay.captured_at < sqlc.arg('to_at')::timestamptz
  GROUP BY 1
),
refunded AS (
  SELECT
    date_trunc(sqlc.arg('grouping')::text, pay.refunded_at AT TIME ZONE sqlc.arg('zone')::text)::date AS start,
    sum(pay.amount_cents) AS refunds_cents
  FROM payments pay
  WHERE pay.refunded_at >= sqlc.arg('from_at')::timestamptz
    AND pay.refunded_at < sqlc.arg('to_at')::timestamptz
  GROUP BY 1
)
SELECT
  GREATEST(p.start, sqlc.arg('from_day')::date)::date AS start,
  COALESCE(t.orders, 0)::bigint AS orders,
  COALESCE(t.gross_cents, 0)::bigint AS gross_cents,
  COALESCE(r.refunds_cents, 0)::bigint AS refunds_cents
FROM periods p
LEFT JOIN taken t ON t.start = p.start
LEFT JOIN refunded r ON r.start = p.start
ORDER BY p.start;

-- name: TopSoldItems :many
-- The products and combos sold most, by sales, on the orders paid in
-- [from_at, to_at) and not cancelled: units and line totals, before any
-- order discount, under the name each has now.
SELECT
  i.line_type,
  COALESCE(i.product_id, i.combo_id)::uuid AS item_id,
  COALESCE(p.name, c.name)::text AS name,
  sum(i.quantity)::bigint AS quantity,
  sum(i.line_total_cents)::bigint AS sales_cents
FROM order_items i
JOIN orders o ON o.id = i.order_id
JOIN payments pay ON pay.order_id = o.id
LEFT JOIN products p ON p.id = i.product_id
LEFT JOIN combos c ON c.id = i.combo_id
WHERE pay.captured_at >= sqlc.arg('from_at')::timestamptz
  AND pay.captured_at < sqlc.arg('to_at')::timestamptz
  AND o.status <> 'cancelled'::order_status
GROUP BY i.line_type, COALESCE(i.product_id, i.combo_id), COALESCE(p.name, c.name)
ORDER BY sales_cents DESC, quantity DESC, COALESCE(p.name, c.name)
LIMIT sqlc.arg('max_rows')::int;

-- name: SalesByCategory :many
-- What the orders paid in [from_at, to_at) and not cancelled sold in each
-- product's category now, before any order discount; combos, which have no
-- category, come as one row without one.
SELECT
  cat.id AS category_id,
  cat.name AS category_name,
  sum(i.quantity)::bigint AS quantity,
  sum(i.line_total_cents)::bigint AS sales_cents
FROM order_items i
JOIN orders o ON o.id = i.order_id
JOIN payments pay ON pay.order_id = o.id
LEFT JOIN products p ON p.id = i.product_id
LEFT JOIN categories cat ON cat.id = p.category_id
WHERE pay.captured_at >= sqlc.arg('from_at')::timestamptz
  AND pay.captured_at < sqlc.arg('to_at')::timestamptz
  AND o.status <> 'cancelled'::order_status
GROUP BY cat.id, cat.name
ORDER BY sales_cents DESC, category_name;

-- name: SalesDiscounts :one
-- The orders paid in [from_at, to_at) and not cancelled, how many of them
-- used a discount code, and how much the codes took off.
SELECT
  count(*)::bigint AS orders,
  count(*) FILTER (WHERE o.discount_code_snapshot IS NOT NULL)::bigint AS discounted_orders,
  COALESCE(sum(o.discount_cents), 0)::bigint AS discount_cents
FROM orders o
JOIN payments pay ON pay.order_id = o.id
WHERE pay.captured_at >= sqlc.arg('from_at')::timestamptz
  AND pay.captured_at < sqlc.arg('to_at')::timestamptz
  AND o.status <> 'cancelled'::order_status;

-- name: SalesDiscountCodes :many
-- The discount codes those orders used: how many orders each, and how much it
-- took off, the most first, up to max_rows.
SELECT
  o.discount_code_snapshot::text AS code,
  count(*)::bigint AS orders,
  sum(o.discount_cents)::bigint AS discount_cents
FROM orders o
JOIN payments pay ON pay.order_id = o.id
WHERE pay.captured_at >= sqlc.arg('from_at')::timestamptz
  AND pay.captured_at < sqlc.arg('to_at')::timestamptz
  AND o.status <> 'cancelled'::order_status
  AND o.discount_code_snapshot IS NOT NULL
GROUP BY o.discount_code_snapshot
ORDER BY discount_cents DESC, code
LIMIT sqlc.arg('max_rows')::int;

-- name: ProductionTime :one
-- The orders that became ready in [from_at, to_at), and on average how many
-- whole minutes they took from the first ticket started (in production) to
-- ready; 0 when none did.
SELECT
  count(*)::bigint AS orders,
  COALESCE(round(avg(extract(epoch FROM r.created_at - s.created_at)) / 60), 0)::int AS average_minutes
FROM order_status_events r
JOIN order_status_events s ON s.order_id = r.order_id AND s.to_status = 'in_production'::order_status
WHERE r.to_status = 'ready'::order_status
  AND r.created_at >= sqlc.arg('from_at')::timestamptz
  AND r.created_at < sqlc.arg('to_at')::timestamptz;
