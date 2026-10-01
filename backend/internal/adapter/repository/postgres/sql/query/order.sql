-- name: CreateOrder :one
INSERT INTO orders (
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  pickup_at,
  code,
  order_type,
  terms_version,
  terms_accepted_at,
  checkout_key,
  payment_due_at
)
VALUES (
  sqlc.arg('user_id'),
  sqlc.arg('status'),
  sqlc.arg('subtotal_cents'),
  sqlc.arg('discount_cents'),
  sqlc.arg('total_cents'),
  sqlc.narg('discount_code_id'),
  sqlc.narg('discount_code_snapshot'),
  sqlc.narg('pickup_at'),
  sqlc.arg('code'),
  sqlc.arg('order_type'),
  sqlc.narg('terms_version')::text,
  -- Accepted at the instant the order is placed: the transaction clock that
  -- created_at takes.
  CASE WHEN sqlc.narg('terms_version')::text IS NULL THEN NULL ELSE now() END,
  sqlc.narg('checkout_key'),
  -- Held unpaid until then, on the database clock.
  now() + make_interval(mins => sqlc.arg('payment_hold_minutes')::int)
)
RETURNING
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  pickup_at,
  created_at,
  updated_at,
  code,
  order_type,
  terms_accepted_at,
  terms_version,
  checkout_key,
  payment_due_at;

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
  pickup_at,
  created_at,
  updated_at,
  code,
  order_type,
  terms_accepted_at,
  terms_version,
  checkout_key,
  payment_due_at
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
  pickup_at,
  created_at,
  updated_at,
  code,
  order_type,
  terms_accepted_at,
  terms_version,
  checkout_key,
  payment_due_at
FROM orders
WHERE user_id = sqlc.arg('user_id')
  AND (
    sqlc.narg('status')::order_status IS NULL
    OR status = sqlc.narg('status')::order_status
  )
  AND (
    sqlc.narg('placed_from')::timestamptz IS NULL
    OR created_at >= sqlc.narg('placed_from')::timestamptz
  )
  AND (
    sqlc.narg('placed_before')::timestamptz IS NULL
    OR created_at < sqlc.narg('placed_before')::timestamptz
  )
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListOrdersByUserBefore :many
-- A customer's orders newest first, a page at a time by keyset: each page
-- starts below the last order of the page before, so an order placed while a
-- data export reads never repeats or hides one across a page boundary.
SELECT
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  pickup_at,
  created_at,
  updated_at,
  code,
  order_type,
  terms_accepted_at,
  terms_version,
  checkout_key,
  payment_due_at
FROM orders
WHERE user_id = sqlc.arg('user_id')
  AND (
    sqlc.narg('before_at')::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg('before_at')::timestamptz, sqlc.narg('before_id')::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit');

-- name: HasOpenOrdersForUser :one
-- Whether the customer has an order the bakery still has to make or hand
-- over, or one waiting for payment: anything not fulfilled, cancelled or
-- expired.
SELECT EXISTS (
  SELECT 1
  FROM orders
  WHERE user_id = $1
    AND status NOT IN ('fulfilled'::order_status, 'cancelled'::order_status, 'expired'::order_status, 'no_show'::order_status)
) AS open;

-- name: ListOrdersByUserCount :one
SELECT COUNT(*)::bigint AS count
FROM orders
WHERE user_id = sqlc.arg('user_id')
  AND (
    sqlc.narg('status')::order_status IS NULL
    OR status = sqlc.narg('status')::order_status
  )
  AND (
    sqlc.narg('placed_from')::timestamptz IS NULL
    OR created_at >= sqlc.narg('placed_from')::timestamptz
  )
  AND (
    sqlc.narg('placed_before')::timestamptz IS NULL
    OR created_at < sqlc.narg('placed_before')::timestamptz
  );

-- name: CreateOrderItems :execrows
-- One round trip for every checkout line: the items arrive as a JSON array and
-- Postgres casts each field to the column type (constraints still apply per row).
INSERT INTO order_items (
  order_id,
  line_type,
  product_id,
  combo_id,
  configuration,
  name,
  slug,
  quantity,
  unit_price_cents,
  line_total_cents
)
SELECT
  item.order_id,
  item.line_type,
  item.product_id,
  item.combo_id,
  item.configuration,
  item.name,
  item.slug,
  item.quantity,
  item.unit_price_cents,
  item.line_total_cents
FROM jsonb_to_recordset(sqlc.arg('items')::jsonb) AS item (
  order_id uuid,
  line_type line_type,
  product_id uuid,
  combo_id uuid,
  configuration jsonb,
  name text,
  slug text,
  quantity integer,
  unit_price_cents bigint,
  line_total_cents bigint
);

-- name: OrderHasCustomItems :one
-- Whether the order holds an item its customer configured, which staff review
-- before the bakery makes it.
SELECT EXISTS (
  SELECT 1 FROM order_items
  WHERE order_id = $1 AND configuration <> '{}'::jsonb
)::bool AS has_custom_items;

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
  configuration,
  name,
  slug,
  quantity,
  unit_price_cents,
  line_total_cents,
  created_at
FROM order_items
WHERE order_id = $1
ORDER BY created_at ASC;

-- name: ListOrderItemsByOrderIDs :many
-- The lines of many orders in one round trip, as a data export reads them.
SELECT
  id,
  order_id,
  line_type,
  product_id,
  combo_id,
  configuration,
  name,
  slug,
  quantity,
  unit_price_cents,
  line_total_cents,
  created_at
FROM order_items
WHERE order_id = ANY(sqlc.arg('order_ids')::uuid[])
ORDER BY order_id, created_at ASC;

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
  o.pickup_at,
  o.created_at,
  o.updated_at,
  o.code,
  o.order_type,
  u.email AS customer_email,
  (u.email_verified_at IS NOT NULL)::boolean AS customer_email_verified,
  cp.display_name AS customer_display_name
FROM orders o
INNER JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
LEFT JOIN customer_profiles cp ON cp.user_id = o.user_id
-- An order not paid, now or ever, is not the bakery's to see: one cancelled
-- is only if it was accepted before (domainorder.SeenByStaff).
WHERE o.status NOT IN ('awaiting_payment'::order_status, 'expired'::order_status)
  AND (o.status <> 'cancelled'::order_status OR EXISTS (
    SELECT 1 FROM order_status_events e
    WHERE e.order_id = o.id AND e.to_status IN ('pending'::order_status, 'confirmed'::order_status)
  ))
  AND (
    sqlc.narg('status')::order_status IS NULL
    OR o.status = sqlc.narg('status')::order_status
  )
ORDER BY o.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: StaffListOrdersCount :one
SELECT COUNT(*)::bigint AS count
FROM orders o
INNER JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
WHERE o.status NOT IN ('awaiting_payment'::order_status, 'expired'::order_status)
  AND (o.status <> 'cancelled'::order_status OR EXISTS (
    SELECT 1 FROM order_status_events e
    WHERE e.order_id = o.id AND e.to_status IN ('pending'::order_status, 'confirmed'::order_status)
  ))
  AND (
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
  o.pickup_at,
  o.created_at,
  o.updated_at,
  o.code,
  o.order_type,
  u.email AS customer_email,
  (u.email_verified_at IS NOT NULL)::boolean AS customer_email_verified,
  cp.display_name AS customer_display_name,
  cp.phone AS customer_phone
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
  pickup_at,
  created_at,
  updated_at,
  code,
  order_type,
  terms_accepted_at,
  terms_version,
  checkout_key,
  payment_due_at;

-- name: NextOrderDayNumber :one
-- The day is read from the transaction's clock, the instant orders.created_at
-- takes, so a code and its created_at always name the same bakery day. The row
-- lock taken by the upsert holds every other checkout on that day until this
-- transaction ends, so numbers are unique and follow each other.
INSERT INTO order_day_counters (day, last_number)
VALUES ((now() AT TIME ZONE sqlc.arg('zone')::text)::date, 1)
ON CONFLICT (day) DO UPDATE
SET last_number = order_day_counters.last_number + 1
RETURNING day, last_number;

-- name: CreateOrderStatusEvent :exec
INSERT INTO order_status_events (order_id, from_status, to_status, actor_id, actor_role, reason)
VALUES (
  sqlc.arg('order_id'),
  sqlc.narg('from_status')::order_status,
  sqlc.arg('to_status')::order_status,
  sqlc.narg('actor_id'),
  sqlc.narg('actor_role')::user_role,
  sqlc.narg('reason')
);

-- name: ListOrderStatusEvents :many
SELECT to_status, actor_role, reason, created_at
FROM order_status_events
WHERE order_id = $1
ORDER BY created_at ASC, id ASC;

-- name: ListOrderStatusEventsByOrderIDs :many
-- The history of many orders in one round trip, as a data export reads it.
SELECT order_id, to_status, actor_role, reason, created_at
FROM order_status_events
WHERE order_id = ANY(sqlc.arg('order_ids')::uuid[])
ORDER BY order_id, created_at ASC, id ASC;

-- name: LockPickupSlot :exec
-- Holds the slot starting at starts_at until the transaction ends, so two
-- checkouts cannot both take its last place. The key is the slot's minute since
-- the epoch, under a namespace of its own.
SELECT pg_advisory_xact_lock(
  sqlc.arg('namespace')::int,
  (extract(epoch FROM sqlc.arg('starts_at')::timestamptz) / 60)::int
);

-- name: CountOrdersInSlot :one
-- Orders holding the slot [from_at, to_at): every order not cancelled or
-- expired whose pickup falls in it, wherever in it an earlier slot grid put
-- that pickup.
SELECT count(*)::bigint AS count
FROM orders
WHERE pickup_at >= sqlc.arg('from_at')::timestamptz
  AND pickup_at < sqlc.arg('to_at')::timestamptz
  AND status NOT IN ('cancelled'::order_status, 'expired'::order_status);

-- name: LockCustomerBookings :exec
-- Holds the customer's bookings until the transaction ends, so two of their
-- checkouts or pickup moves cannot both take a day's last place. The key is the
-- customer, under a namespace of its own.
SELECT pg_advisory_xact_lock(sqlc.arg('namespace')::int, hashtext(sqlc.arg('user_id')::uuid::text));

-- name: CountCustomerOrdersBetween :one
-- A customer's orders not cancelled or expired with a pickup in [from_at, to_at).
SELECT count(*)::bigint AS count
FROM orders
WHERE user_id = sqlc.arg('user_id')
  AND pickup_at >= sqlc.arg('from_at')::timestamptz
  AND pickup_at < sqlc.arg('to_at')::timestamptz
  AND status NOT IN ('cancelled'::order_status, 'expired'::order_status);

-- name: CountOrdersByPickupTime :many
SELECT pickup_at::timestamptz AS pickup_at, count(*)::bigint AS count
FROM orders
WHERE pickup_at >= sqlc.arg('from_at')::timestamptz
  AND pickup_at < sqlc.arg('to_at')::timestamptz
  AND status NOT IN ('cancelled'::order_status, 'expired'::order_status)
GROUP BY pickup_at;

-- name: LockOrder :one
-- Holds the order row until the transaction ends: every ticket move takes it
-- first, and the order's own status moves take the same row lock through their
-- guarded UPDATE, so its status is derived from tickets no one else is moving.
SELECT
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  pickup_at,
  created_at,
  updated_at,
  code,
  order_type,
  terms_accepted_at,
  terms_version,
  checkout_key,
  payment_due_at
FROM orders
WHERE id = $1
FOR UPDATE;

-- name: GetOrderByCheckoutKey :one
SELECT
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  pickup_at,
  created_at,
  updated_at,
  code,
  order_type,
  terms_accepted_at,
  terms_version,
  checkout_key,
  payment_due_at
FROM orders
WHERE user_id = sqlc.arg('user_id')
  AND checkout_key = sqlc.arg('checkout_key');

-- name: CountCustomerDiscountUses :one
-- A customer's orders not cancelled or expired that use the discount code.
SELECT count(*)::bigint AS count
FROM orders
WHERE user_id = sqlc.arg('user_id')
  AND discount_code_id = sqlc.arg('discount_code_id')
  AND status NOT IN ('cancelled'::order_status, 'expired'::order_status);

-- name: ListDueUnpaidOrders :many
-- Orders awaiting payment more than grace_seconds past their due time, the
-- longest overdue first.
SELECT id
FROM orders
WHERE status = 'awaiting_payment'::order_status
  AND payment_due_at <= now() - make_interval(secs => sqlc.arg('grace_seconds')::double precision)
ORDER BY payment_due_at
LIMIT sqlc.arg('max_rows')::int;

-- name: ExpireOrder :one
-- An order still awaiting payment more than grace_seconds past its due time
-- expires; one paid meanwhile does not.
UPDATE orders
SET status = 'expired'::order_status,
    updated_at = now()
WHERE id = sqlc.arg('id')
  AND status = 'awaiting_payment'::order_status
  AND payment_due_at <= now() - make_interval(secs => sqlc.arg('grace_seconds')::double precision)
RETURNING
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  pickup_at,
  created_at,
  updated_at,
  code,
  order_type,
  terms_accepted_at,
  terms_version,
  checkout_key,
  payment_due_at;

-- name: RescheduleOrder :one
-- Moves the pickup of an order not being made yet; its type follows the new
-- time.
UPDATE orders
SET pickup_at  = sqlc.arg('pickup_at'),
    order_type = sqlc.arg('order_type'),
    updated_at = now()
WHERE id = sqlc.arg('id')
  AND status IN ('awaiting_payment'::order_status, 'pending'::order_status, 'confirmed'::order_status)
RETURNING
  id,
  user_id,
  status,
  subtotal_cents,
  discount_cents,
  total_cents,
  discount_code_id,
  discount_code_snapshot,
  pickup_at,
  created_at,
  updated_at,
  code,
  order_type,
  terms_accepted_at,
  terms_version,
  checkout_key,
  payment_due_at;

-- name: ListMissedPickups :many
-- Orders ready and not collected whose pickup falls before missed_before, the
-- longest waiting first.
SELECT id
FROM orders
WHERE status = 'ready'::order_status
  AND pickup_at < sqlc.arg('missed_before')
ORDER BY pickup_at
LIMIT sqlc.arg('max_rows')::int;
