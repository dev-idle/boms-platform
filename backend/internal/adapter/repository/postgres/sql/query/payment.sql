-- name: CreatePayment :one
-- One payment per order: when the order already has one, nothing is written
-- and no row comes back.
INSERT INTO payments (order_id, provider, provider_order_id, approve_url, amount_cents, currency)
VALUES (
  sqlc.arg('order_id'),
  sqlc.arg('provider'),
  sqlc.arg('provider_order_id')::text,
  sqlc.arg('approve_url')::text,
  sqlc.arg('amount_cents'),
  sqlc.arg('currency')
)
ON CONFLICT (order_id) DO NOTHING
RETURNING *;

-- name: CreateCashPayment :exec
-- Cash due when the order is collected.
INSERT INTO payments (order_id, provider, amount_cents, currency)
VALUES (sqlc.arg('order_id'), 'cash'::payment_provider, sqlc.arg('amount_cents'), sqlc.arg('currency'));

-- name: CollectCashPayment :execrows
-- The counter took the cash due as it handed the order over.
UPDATE payments
SET status      = 'captured',
    captured_at = now(),
    updated_at  = now()
WHERE order_id = sqlc.arg('order_id')
  AND provider = 'cash'::payment_provider
  AND status = 'created';

-- name: GetPaymentByOrderID :one
SELECT *
FROM payments
WHERE order_id = $1;

-- name: GetPaymentByProviderOrderID :one
SELECT *
FROM payments
WHERE provider = sqlc.arg('provider')
  AND provider_order_id = sqlc.arg('provider_order_id')::text;

-- name: RestartPayment :one
-- A payment the provider denied starts over with a new provider order.
UPDATE payments
SET provider_order_id = sqlc.arg('provider_order_id')::text,
    approve_url       = sqlc.arg('approve_url')::text,
    status            = 'created',
    capture_id        = NULL,
    updated_at        = now()
WHERE order_id = sqlc.arg('order_id')
  AND status = 'denied'
RETURNING *;

-- name: RecordPaymentCapture :one
-- Records the provider's answer to a capture: taken, held for review, or a
-- held one refused. A payment already taken or refused stays as it is.
UPDATE payments
SET status      = sqlc.arg('status'),
    capture_id  = sqlc.arg('capture_id'),
    captured_at = CASE WHEN sqlc.arg('status')::payment_status = 'captured' THEN now() END,
    updated_at  = now()
WHERE id = sqlc.arg('id')
  AND status IN ('created', 'pending')
RETURNING *;

-- name: RequestPaymentRefund :exec
-- Asks for the money back on a cancelled order's payment, in the transaction
-- that cancels it. A payment not taken, or asked already, stays as it is.
UPDATE payments
SET refund_requested_at = now(),
    updated_at          = now()
WHERE order_id = $1
  AND status = 'captured'
  AND refund_requested_at IS NULL;

-- name: ListRefundsDue :many
-- Payments whose refund is asked for and not made yet, the oldest first.
SELECT *
FROM payments
WHERE status = 'captured'
  AND refund_requested_at IS NOT NULL
ORDER BY refund_requested_at
LIMIT sqlc.arg('max_rows')::int;

-- name: RecordPaymentRefund :one
-- Records the provider's refund of a payment whose refund was asked for; a
-- payment refunded already stays as it is.
UPDATE payments
SET status      = 'refunded',
    refund_id   = sqlc.narg('refund_id'),
    refunded_at = now(),
    updated_at  = now()
WHERE id = sqlc.arg('id')
  AND status = 'captured'
  AND refund_requested_at IS NOT NULL
RETURNING *;
