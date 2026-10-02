-- name: CountEngagement :one
-- How many open customer accounts there are, how many of them used each
-- engagement feature, and how many used none, one, or two or more. Reviews
-- and messages count what the customer wrote; favorites and wishlist count
-- every product ever saved, removed ones included; promotions count those
-- agreeing to them now. The usage is worked out once per customer
-- (MATERIALIZED, so the counts below do not repeat it), each feature's
-- customers read once as a distinct set.
WITH usage AS MATERIALIZED (
  SELECT
    u.id IN (SELECT DISTINCT user_id FROM reviews WHERE deleted_at IS NULL) AS reviews,
    u.id IN (SELECT DISTINCT user_id FROM saved_products WHERE list = 'favorite') AS favorites,
    u.id IN (SELECT DISTINCT user_id FROM saved_products WHERE list = 'wishlist') AS wishlist,
    u.id IN (SELECT DISTINCT author_id FROM messages WHERE author_role = 'customer' AND deleted_at IS NULL) AS messages,
    u.id IN (SELECT user_id FROM customer_profiles WHERE marketing_consent_at IS NOT NULL) AS promotions
  FROM users u
  WHERE u.role = 'customer'
    AND u.deleted_at IS NULL
),
features AS (
  SELECT
    usage.*,
    reviews::int + favorites::int + wishlist::int + messages::int + promotions::int AS used
  FROM usage
)
SELECT
  count(*)::bigint AS customers,
  count(*) FILTER (WHERE reviews)::bigint AS reviews,
  count(*) FILTER (WHERE favorites)::bigint AS favorites,
  count(*) FILTER (WHERE wishlist)::bigint AS wishlist,
  count(*) FILTER (WHERE messages)::bigint AS messages,
  count(*) FILTER (WHERE promotions)::bigint AS promotions,
  count(*) FILTER (WHERE used = 0)::bigint AS used_none,
  count(*) FILTER (WHERE used = 1)::bigint AS used_one,
  count(*) FILTER (WHERE used >= 2)::bigint AS used_two_or_more
FROM features;
