-- name: CreateReview :one
INSERT INTO reviews (order_id, product_id, user_id, rating, comment)
VALUES (sqlc.arg('order_id'), sqlc.arg('product_id'), sqlc.arg('user_id'), sqlc.arg('rating'), sqlc.narg('comment'))
RETURNING id, product_id, rating, comment, status, created_at;

-- name: ListReviewsByOrder :many
-- The customer's reviews of the products on one of their orders.
SELECT id, product_id, rating, comment, status, created_at
FROM reviews
WHERE order_id = sqlc.arg('order_id')
  AND user_id = sqlc.arg('user_id')
  AND deleted_at IS NULL
ORDER BY created_at, id;

-- name: ListPublishedReviews :many
-- A page of a product's published reviews, latest first: the latest, or those
-- written before the review before_id names.
SELECT r.id, r.rating, r.comment, r.created_at
FROM reviews r
WHERE r.product_id = sqlc.arg('product_id')
  AND r.status = 'published'::review_status
  AND r.deleted_at IS NULL
  AND (
    sqlc.narg('before_id')::uuid IS NULL
    OR (r.created_at, r.id) < (
      SELECT b.created_at, b.id FROM reviews b
      WHERE b.id = sqlc.narg('before_id')::uuid AND b.product_id = r.product_id
    )
  )
ORDER BY r.created_at DESC, r.id DESC
LIMIT sqlc.arg('limit');

-- name: CountPublishedReviews :one
-- How many published reviews a product has, and their ratings added up.
SELECT
  count(*)::bigint AS review_count,
  COALESCE(sum(rating), 0)::bigint AS rating_total
FROM reviews
WHERE product_id = sqlc.arg('product_id')
  AND status = 'published'::review_status
  AND deleted_at IS NULL;

-- name: ManagerListReviews :many
-- A page of reviews to moderate, latest first, with the product, the order
-- and the name the customer gave.
SELECT
  r.id,
  o.code AS order_code,
  r.product_id,
  p.name AS product_name,
  r.rating,
  r.comment,
  r.status,
  r.created_at,
  cp.display_name AS customer_name,
  sp.full_name AS moderator_name
FROM reviews r
JOIN orders o ON o.id = r.order_id
JOIN products p ON p.id = r.product_id
LEFT JOIN customer_profiles cp ON cp.user_id = r.user_id
LEFT JOIN staff_profiles sp ON sp.user_id = r.moderated_by
WHERE r.deleted_at IS NULL
  AND (
    sqlc.narg('status')::review_status IS NULL
    OR r.status = sqlc.narg('status')::review_status
  )
ORDER BY r.created_at DESC, r.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ManagerListReviewsCount :one
SELECT count(*)::bigint
FROM reviews r
WHERE r.deleted_at IS NULL
  AND (
    sqlc.narg('status')::review_status IS NULL
    OR r.status = sqlc.narg('status')::review_status
  );

-- name: GetReviewForUpdate :one
-- Where a review stands and who wrote it, held until the transaction ends.
SELECT status, user_id
FROM reviews
WHERE id = $1
  AND deleted_at IS NULL
FOR UPDATE;

-- name: ModerateReview :one
-- Publishes or hides a review, and returns it as the moderation list shows it.
WITH moderated AS (
  UPDATE reviews
  SET status = sqlc.arg('status'),
      moderated_by = sqlc.arg('moderated_by'),
      moderated_at = now(),
      updated_at = now()
  WHERE reviews.id = sqlc.arg('id')
    AND reviews.deleted_at IS NULL
  RETURNING id, order_id, product_id, user_id, rating, comment, status, created_at, moderated_by
)
SELECT
  m.id,
  o.code AS order_code,
  m.product_id,
  p.name AS product_name,
  m.rating,
  m.comment,
  m.status,
  m.created_at,
  cp.display_name AS customer_name,
  sp.full_name AS moderator_name
FROM moderated m
JOIN orders o ON o.id = m.order_id
JOIN products p ON p.id = m.product_id
LEFT JOIN customer_profiles cp ON cp.user_id = m.user_id
LEFT JOIN staff_profiles sp ON sp.user_id = m.moderated_by;

-- name: ReviewRatingCounts :one
-- Every review not hidden of a product the bakery keeps: how many, how many
-- wait for a manager, their ratings added up, and how many of each rating.
SELECT
  count(*)::bigint AS review_count,
  count(*) FILTER (WHERE r.status = 'pending'::review_status)::bigint AS pending_count,
  COALESCE(sum(r.rating), 0)::bigint AS rating_total,
  count(*) FILTER (WHERE r.rating = 1)::bigint AS one_star,
  count(*) FILTER (WHERE r.rating = 2)::bigint AS two_stars,
  count(*) FILTER (WHERE r.rating = 3)::bigint AS three_stars,
  count(*) FILTER (WHERE r.rating = 4)::bigint AS four_stars,
  count(*) FILTER (WHERE r.rating = 5)::bigint AS five_stars
FROM reviews r
JOIN products p ON p.id = r.product_id AND p.deleted_at IS NULL
WHERE r.deleted_at IS NULL
  AND r.status <> 'hidden'::review_status;

-- name: ReviewRatingsByProduct :many
-- Each reviewed product the bakery keeps, its reviews not hidden added up, the
-- lowest average first: what needs a manager's attention leads.
SELECT
  p.id,
  p.name,
  count(*)::bigint AS review_count,
  sum(r.rating)::bigint AS rating_total
FROM reviews r
JOIN products p ON p.id = r.product_id AND p.deleted_at IS NULL
WHERE r.deleted_at IS NULL
  AND r.status <> 'hidden'::review_status
GROUP BY p.id, p.name
ORDER BY sum(r.rating)::numeric / count(*), count(*) DESC, p.name
LIMIT sqlc.arg('limit');

-- name: ListReviewsByUser :many
-- Every review the customer wrote, for a personal data export.
SELECT
  r.order_id,
  o.code AS order_code,
  r.product_id,
  p.name AS product_name,
  r.rating,
  r.comment,
  r.status,
  r.created_at
FROM reviews r
JOIN orders o ON o.id = r.order_id
JOIN products p ON p.id = r.product_id
WHERE r.user_id = sqlc.arg('user_id')
  AND r.deleted_at IS NULL
ORDER BY r.created_at, r.id;

-- name: EraseUserReviews :exec
-- Erases what the customer wrote as their account is erased; their reviews
-- no longer count.
UPDATE reviews
SET comment = NULL,
    deleted_at = now(),
    updated_at = now()
WHERE user_id = sqlc.arg('user_id')
  AND deleted_at IS NULL;
