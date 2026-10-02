-- name: OperationsActiveOrders :one
-- The orders the bakery still has to make or hand over, by status; like the
-- counter's and the kitchen's lists, none of a closed account.
SELECT
  count(*) FILTER (WHERE o.status = 'pending'::order_status)::bigint AS pending,
  count(*) FILTER (WHERE o.status = 'confirmed'::order_status)::bigint AS confirmed,
  count(*) FILTER (WHERE o.status = 'in_production'::order_status)::bigint AS in_production,
  count(*) FILTER (WHERE o.status = 'ready'::order_status)::bigint AS ready
FROM orders o
LEFT JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
WHERE (o.user_id IS NULL OR u.id IS NOT NULL)
  AND o.status IN ('pending', 'confirmed', 'in_production', 'ready');

-- name: OperationsStationLoads :many
-- The tickets of the orders being made or waiting at the counter, at each
-- station, by status; a station with none is left out.
SELECT
  t.station,
  count(*) FILTER (WHERE t.status = 'queued'::ticket_status)::bigint AS queued,
  count(*) FILTER (WHERE t.status = 'in_progress'::ticket_status)::bigint AS in_progress,
  count(*) FILTER (WHERE t.status = 'ready'::ticket_status)::bigint AS ready
FROM order_tickets t
JOIN orders o ON o.id = t.order_id
LEFT JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
WHERE (o.user_id IS NULL OR u.id IS NOT NULL)
  AND o.status IN ('confirmed', 'in_production', 'ready')
  AND t.status <> 'cancelled'::ticket_status
GROUP BY t.station
ORDER BY t.station;

-- name: OperationsDayPickups :one
-- The orders the bakery took with a pickup in [from_at, to_at): how many,
-- collected, still to collect, and not collected.
SELECT
  count(*)::bigint AS due,
  count(*) FILTER (WHERE o.status = 'fulfilled'::order_status)::bigint AS collected,
  count(*) FILTER (WHERE o.status IN ('pending', 'confirmed', 'in_production', 'ready'))::bigint AS to_collect,
  count(*) FILTER (WHERE o.status = 'no_show'::order_status)::bigint AS missed
FROM orders o
LEFT JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
WHERE (o.user_id IS NULL OR u.id IS NOT NULL)
  AND o.pickup_at >= sqlc.arg('from_at')::timestamptz
  AND o.pickup_at < sqlc.arg('to_at')::timestamptz
  AND o.status NOT IN ('awaiting_payment', 'expired', 'cancelled');

-- name: OperationsLateOrders :many
-- The orders still to collect whose pickup slot has ended by now, as the
-- counter marks them late, the longest waiting first, with how many there are.
SELECT
  o.code,
  o.status,
  o.pickup_at::timestamptz AS pickup_at,
  count(*) OVER ()::bigint AS total
FROM orders o
CROSS JOIN store_settings s
LEFT JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
WHERE s.id = 1
  AND (o.user_id IS NULL OR u.id IS NOT NULL)
  AND o.status IN ('pending', 'confirmed', 'in_production', 'ready')
  AND o.pickup_at <= sqlc.arg('now')::timestamptz - make_interval(mins => s.slot_minutes)
ORDER BY o.pickup_at, o.id
LIMIT sqlc.arg('max_rows')::int;
