-- name: SaveProduct :execrows
-- Puts a product on sale on one of the customer's lists, or brings back the
-- row they removed. Saving it again leaves it where it is on the list. No row
-- is written for a product that is not on sale.
INSERT INTO saved_products (user_id, product_id, list)
SELECT sqlc.arg('user_id'), p.id, sqlc.arg('list')
FROM products p
JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL AND c.is_active = true
WHERE p.id = sqlc.arg('product_id')
  AND p.deleted_at IS NULL
  AND p.is_active = true
ON CONFLICT (user_id, product_id, list) DO UPDATE
SET saved_at = CASE WHEN saved_products.deleted_at IS NULL THEN saved_products.saved_at ELSE now() END,
    deleted_at = NULL,
    updated_at = now();

-- name: CountOtherSavedProducts :one
-- How many products on sale the customer's list shows besides the one given.
SELECT count(*)::bigint
FROM saved_products s
JOIN products p ON p.id = s.product_id AND p.deleted_at IS NULL AND p.is_active = true
JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL AND c.is_active = true
WHERE s.user_id = sqlc.arg('user_id')
  AND s.list = sqlc.arg('list')
  AND s.product_id <> sqlc.arg('product_id')
  AND s.deleted_at IS NULL;

-- name: RemoveSavedProduct :exec
UPDATE saved_products
SET deleted_at = now(),
    updated_at = now()
WHERE user_id = sqlc.arg('user_id')
  AND product_id = sqlc.arg('product_id')
  AND list = sqlc.arg('list')
  AND deleted_at IS NULL;

-- name: ListSavedProducts :many
-- Both of the customer's lists, latest first, with each product as the
-- catalog shows it. A product the bakery stopped selling is kept but not
-- shown, as the catalog does not show it; it is back on the list when it is
-- back on sale.
SELECT
  s.list,
  s.saved_at,
  p.id,
  p.category_id,
  p.name,
  p.slug,
  p.price_cents,
  p.is_customizable,
  p.sold_out_on,
  c.name AS category_name,
  c.slug AS category_slug
FROM saved_products s
JOIN products p ON p.id = s.product_id AND p.deleted_at IS NULL AND p.is_active = true
JOIN categories c ON c.id = p.category_id AND c.deleted_at IS NULL AND c.is_active = true
WHERE s.user_id = sqlc.arg('user_id')
  AND s.deleted_at IS NULL
ORDER BY s.list, s.saved_at DESC, s.id DESC;

-- name: ListSavedProductsForExport :many
-- Every product the customer saved to a list, shown or not, taken off or not
-- (with when), for a personal data export: what is held about them, whatever
-- the bakery does with the product.
SELECT
  s.list,
  s.saved_at,
  s.deleted_at AS removed_at,
  p.id,
  p.name
FROM saved_products s
JOIN products p ON p.id = s.product_id
WHERE s.user_id = sqlc.arg('user_id')
ORDER BY s.list, s.saved_at DESC, s.id DESC;

-- name: RemoveAllSavedProducts :exec
-- Empties the customer's lists as their account is erased.
UPDATE saved_products
SET deleted_at = now(),
    updated_at = now()
WHERE user_id = sqlc.arg('user_id')
  AND deleted_at IS NULL;
