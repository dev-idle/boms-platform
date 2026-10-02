-- name: CreatePromotion :one
-- Records a promotion to send, and returns it as managers list it.
WITH created AS (
  INSERT INTO promotions (subject, body, created_by)
  VALUES (sqlc.arg('subject'), sqlc.arg('body'), sqlc.arg('created_by'))
  RETURNING id, subject, status, recipient_count, created_at, created_by
)
SELECT
  c.id,
  c.subject,
  c.status,
  c.recipient_count,
  c.created_at,
  sp.full_name AS sender_name
FROM created c
LEFT JOIN staff_profiles sp ON sp.user_id = c.created_by;

-- name: ListPromotions :many
-- A page of the promotions sent, latest first, with the manager who sent each.
SELECT
  p.id,
  p.subject,
  p.status,
  p.recipient_count,
  p.created_at,
  sp.full_name AS sender_name
FROM promotions p
LEFT JOIN staff_profiles sp ON sp.user_id = p.created_by
ORDER BY p.created_at DESC, p.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountPromotions :one
SELECT count(*)::bigint FROM promotions;

-- name: GetPromotionStatus :one
SELECT status
FROM promotions
WHERE id = $1;

-- name: ListPromotionRecipients :many
-- A page of the customers promotions go to, by id after after_id: open
-- accounts with a confirmed address whose holder agreed to promotions.
SELECT cp.user_id
FROM customer_profiles cp
JOIN users u ON u.id = cp.user_id
WHERE cp.marketing_consent_at IS NOT NULL
  AND u.deleted_at IS NULL
  AND u.email_verified_at IS NOT NULL
  AND u.role = 'customer'::user_role
  AND (sqlc.narg('after_id')::uuid IS NULL OR cp.user_id > sqlc.narg('after_id')::uuid)
ORDER BY cp.user_id
LIMIT sqlc.arg('limit');

-- name: CountPromotionRecipients :one
SELECT count(*)::bigint
FROM customer_profiles cp
JOIN users u ON u.id = cp.user_id
WHERE cp.marketing_consent_at IS NOT NULL
  AND u.deleted_at IS NULL
  AND u.email_verified_at IS NOT NULL
  AND u.role = 'customer'::user_role;

-- name: GetPromotionDelivery :one
-- What one customer's email of a promotion says and where it goes, while
-- promotions still go to them.
SELECT p.subject, p.body, u.email, cp.display_name
FROM promotions p
CROSS JOIN customer_profiles cp
JOIN users u ON u.id = cp.user_id
WHERE p.id = sqlc.arg('promotion_id')
  AND cp.user_id = sqlc.arg('user_id')
  AND cp.marketing_consent_at IS NOT NULL
  AND u.deleted_at IS NULL
  AND u.email_verified_at IS NOT NULL
  AND u.role = 'customer'::user_role;

-- name: MarkPromotionSent :execrows
-- Records how many customers a promotion was queued for, once.
UPDATE promotions
SET status = 'sent'::promotion_status,
    recipient_count = sqlc.arg('recipient_count'),
    updated_at = now()
WHERE id = sqlc.arg('id')
  AND status = 'sending'::promotion_status;
