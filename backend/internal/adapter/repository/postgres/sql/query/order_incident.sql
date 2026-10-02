-- name: CreateOrderIncident :one
INSERT INTO order_incidents (order_id, type, note, actor_id, actor_role)
VALUES (
  sqlc.arg('order_id'),
  sqlc.arg('type')::order_incident_type,
  sqlc.narg('note'),
  sqlc.narg('actor_id'),
  sqlc.narg('actor_role')::user_role
)
RETURNING id, order_id, type, note, actor_id, actor_role, created_at;

-- name: LockCustomerIncidents :exec
-- Holds the flags of the order's customer until the transaction ends, so two
-- of their payments failing at once both count. The key is the customer, under
-- a namespace of its own; an order without a customer locks nothing.
SELECT pg_advisory_xact_lock(sqlc.arg('namespace')::int, hashtext(user_id::text))
FROM orders
WHERE id = sqlc.arg('order_id')
  AND user_id IS NOT NULL;

-- name: FlagPaymentAnomaly :one
-- Flags the customer of an order when their orders' payments failed or expired
-- threshold times within the window, unless a flag within it already covers
-- them; no row when it records none. An order without a customer flags nobody.
WITH recent AS (
  SELECT i.type
  FROM order_incidents i
  JOIN orders o ON o.id = i.order_id
  JOIN orders flagged ON flagged.user_id = o.user_id
  WHERE flagged.id = sqlc.arg('order_id')
    AND i.type IN ('payment_failed', 'payment_expired', 'payment_anomaly')
    AND i.created_at > now() - make_interval(secs => sqlc.arg('window_seconds')::double precision)
)
INSERT INTO order_incidents (order_id, type)
SELECT sqlc.arg('order_id'), 'payment_anomaly'::order_incident_type
WHERE (SELECT count(*) FROM recent WHERE type <> 'payment_anomaly') >= sqlc.arg('threshold')::int
  AND NOT EXISTS (SELECT 1 FROM recent WHERE type = 'payment_anomaly')
RETURNING id, order_id, type, note, actor_id, actor_role, created_at;

-- name: ListOrderIncidents :many
-- A page of the incidents recorded in [from_at, to_at), latest first, with
-- each order's code and pickup time, the staff member who recorded it and, for
-- a bakery cancellation, the reason the customer was given (an order is
-- cancelled once).
SELECT
  i.id,
  i.order_id,
  i.type,
  i.note,
  i.created_at,
  o.code AS order_code,
  o.pickup_at,
  sp.full_name AS actor_name,
  cancelled.reason AS cancel_reason
FROM order_incidents i
JOIN orders o ON o.id = i.order_id
LEFT JOIN staff_profiles sp ON sp.user_id = i.actor_id
LEFT JOIN order_status_events cancelled
  ON i.type = 'bakery_cancelled'
  AND cancelled.order_id = i.order_id
  AND cancelled.to_status = 'cancelled'
WHERE i.created_at >= sqlc.arg('from_at')
  AND i.created_at < sqlc.arg('to_at')
  AND (sqlc.narg('type')::order_incident_type IS NULL OR i.type = sqlc.narg('type')::order_incident_type)
ORDER BY i.created_at DESC, i.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountOrderIncidents :one
SELECT count(*)::bigint
FROM order_incidents
WHERE created_at >= sqlc.arg('from_at')
  AND created_at < sqlc.arg('to_at')
  AND (sqlc.narg('type')::order_incident_type IS NULL OR type = sqlc.narg('type')::order_incident_type);

-- name: CountOrderIncidentsByType :many
SELECT type, count(*)::bigint AS count
FROM order_incidents
WHERE created_at >= sqlc.arg('from_at')
  AND created_at < sqlc.arg('to_at')
GROUP BY type
ORDER BY type;

-- name: ListOrderIncidentsByOrderIDs :many
-- The incidents of many orders in one round trip, as a data export reads them.
SELECT id, order_id, type, note, actor_id, actor_role, created_at
FROM order_incidents
WHERE order_id = ANY(sqlc.arg('order_ids')::uuid[])
ORDER BY order_id, created_at ASC, id ASC;

-- name: EraseOrderIncidentNotes :exec
UPDATE order_incidents
SET note = NULL
WHERE note IS NOT NULL
  AND order_id IN (SELECT id FROM orders WHERE user_id = sqlc.arg('user_id')::uuid);
