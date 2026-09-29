-- name: ListOrderTicketLines :many
-- What an order's lines ask each station to make: a product line itself, by
-- the name on the receipt, and every product a combo line holds, times the
-- line's quantity.
SELECT
  oi.id AS order_item_id,
  p.id AS product_id,
  oi.name,
  c.station,
  oi.quantity::int AS quantity
FROM order_items oi
INNER JOIN products p ON p.id = oi.product_id
INNER JOIN categories c ON c.id = p.category_id
WHERE oi.order_id = $1
  AND oi.line_type = 'product'
UNION ALL
SELECT
  oi.id AS order_item_id,
  p.id AS product_id,
  p.name,
  c.station,
  (oi.quantity * ci.quantity)::int AS quantity
FROM order_items oi
INNER JOIN combo_items ci ON ci.combo_id = oi.combo_id
INNER JOIN products p ON p.id = ci.product_id
INNER JOIN categories c ON c.id = p.category_id
WHERE oi.order_id = $1
  AND oi.line_type = 'combo';

-- name: CreateOrderTickets :many
-- One statement for every station of an order, like CreateOrderItems.
INSERT INTO order_tickets (order_id, station)
SELECT sqlc.arg('order_id'), ticket.station
FROM jsonb_to_recordset(sqlc.arg('tickets')::jsonb) AS ticket (station station)
RETURNING id, order_id, station, status, created_at, updated_at;

-- name: CreateOrderTicketItems :execrows
INSERT INTO order_ticket_items (ticket_id, order_item_id, product_id, name, quantity)
SELECT item.ticket_id, item.order_item_id, item.product_id, item.name, item.quantity
FROM jsonb_to_recordset(sqlc.arg('items')::jsonb) AS item (
  ticket_id uuid,
  order_item_id uuid,
  product_id uuid,
  name text,
  quantity integer
);

-- name: GetOrderTicket :one
SELECT id, order_id, station, status, created_at, updated_at
FROM order_tickets
WHERE id = $1;

-- name: ListOrderTickets :many
SELECT id, order_id, station, status, created_at, updated_at
FROM order_tickets
WHERE order_id = $1
ORDER BY station;

-- name: ListOrderTicketItems :many
-- The items of every ticket of an order.
SELECT i.ticket_id, i.order_item_id, i.product_id, i.name, i.quantity
FROM order_ticket_items i
INNER JOIN order_tickets t ON t.id = i.ticket_id
WHERE t.order_id = $1
ORDER BY i.created_at, i.name;

-- name: UpdateOrderTicketStatus :one
-- A move made at the ticket's own station; a ticket that already moved on, or
-- sits at another station, is left alone.
UPDATE order_tickets
SET status = sqlc.arg('to_status')::ticket_status,
    updated_at = now()
WHERE id = sqlc.arg('id')
  AND station = sqlc.arg('station')::station
  AND status = sqlc.arg('from_status')::ticket_status
RETURNING id, order_id, station, status, created_at, updated_at;

-- name: MoveOrderTicket :one
-- Only a ticket nobody has started moves; the order's unique ticket per
-- station refuses a station that already has one.
UPDATE order_tickets
SET station = sqlc.arg('to_station')::station,
    updated_at = now()
WHERE id = sqlc.arg('id')
  AND station = sqlc.arg('from_station')::station
  AND status = 'queued'::ticket_status
RETURNING id, order_id, station, status, created_at, updated_at;

-- name: CancelOrderTickets :many
UPDATE order_tickets
SET status = 'cancelled'::ticket_status,
    updated_at = now()
WHERE order_id = $1
  AND status <> 'cancelled'::ticket_status
RETURNING id, order_id, station, status, created_at, updated_at;

-- name: ListStationTickets :many
-- A station's queue: its tickets of orders accepted and not yet collected,
-- soonest pickup first, each with what it makes. The page is cut first and
-- the items gathered after, as in CatalogListProducts.
SELECT
  page.id,
  page.order_id,
  page.station,
  page.status,
  page.created_at,
  page.updated_at,
  page.order_code,
  page.order_status,
  page.pickup_at,
  page.customer_display_name,
  COALESCE(items.list, '[]'::json)::json AS items
FROM (
  SELECT
    t.id,
    t.order_id,
    t.station,
    t.status,
    t.created_at,
    t.updated_at,
    o.code AS order_code,
    o.status AS order_status,
    o.pickup_at,
    cp.display_name AS customer_display_name
  FROM order_tickets t
  INNER JOIN orders o ON o.id = t.order_id
  INNER JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
  LEFT JOIN customer_profiles cp ON cp.user_id = o.user_id
  WHERE t.station = sqlc.arg('station')::station
    AND t.status <> 'cancelled'::ticket_status
    AND o.status IN ('confirmed'::order_status, 'in_production'::order_status, 'ready'::order_status)
    AND (
      sqlc.narg('status')::ticket_status IS NULL
      OR t.status = sqlc.narg('status')::ticket_status
    )
  ORDER BY o.pickup_at ASC NULLS LAST, t.created_at ASC, t.id ASC
  LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset')
) page
LEFT JOIN LATERAL (
  SELECT json_agg(json_build_object('name', i.name, 'quantity', i.quantity) ORDER BY i.created_at, i.name) AS list
  FROM order_ticket_items i
  WHERE i.ticket_id = page.id
) items ON true
ORDER BY page.pickup_at ASC NULLS LAST, page.created_at ASC, page.id ASC;

-- name: CountStationTickets :one
SELECT count(*)::bigint AS count
FROM order_tickets t
INNER JOIN orders o ON o.id = t.order_id
INNER JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
WHERE t.station = sqlc.arg('station')::station
  AND t.status <> 'cancelled'::ticket_status
  AND o.status IN ('confirmed'::order_status, 'in_production'::order_status, 'ready'::order_status)
  AND (
    sqlc.narg('status')::ticket_status IS NULL
    OR t.status = sqlc.narg('status')::ticket_status
  );
