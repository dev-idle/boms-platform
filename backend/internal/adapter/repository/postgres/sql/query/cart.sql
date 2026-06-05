-- name: GetCartByUserID :one
SELECT id, user_id, discount_code_id, created_at, updated_at
FROM carts
WHERE user_id = $1;

-- name: CreateCart :one
INSERT INTO carts (user_id)
VALUES ($1)
RETURNING id, user_id, discount_code_id, created_at, updated_at;

-- name: SetCartDiscountCodeID :exec
UPDATE carts
SET discount_code_id = $2, updated_at = now()
WHERE id = $1;

-- name: ClearCartDiscountCode :exec
UPDATE carts
SET discount_code_id = NULL, updated_at = now()
WHERE id = $1;

-- name: ListCartItemsByCartID :many
SELECT id, cart_id, line_type, product_id, combo_id, quantity, created_at, updated_at
FROM cart_items
WHERE cart_id = $1
ORDER BY created_at ASC;

-- name: GetCartItemByID :one
SELECT id, cart_id, line_type, product_id, combo_id, quantity, created_at, updated_at
FROM cart_items
WHERE cart_id = $1 AND id = $2;

-- name: GetCartItemByProduct :one
SELECT id, cart_id, line_type, product_id, combo_id, quantity, created_at, updated_at
FROM cart_items
WHERE cart_id = $1 AND line_type = 'product' AND product_id = $2;

-- name: GetCartItemByCombo :one
SELECT id, cart_id, line_type, product_id, combo_id, quantity, created_at, updated_at
FROM cart_items
WHERE cart_id = $1 AND line_type = 'combo' AND combo_id = $2;

-- name: CreateCartItem :one
INSERT INTO cart_items (cart_id, line_type, product_id, combo_id, quantity)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, cart_id, line_type, product_id, combo_id, quantity, created_at, updated_at;

-- name: UpdateCartItemQuantity :one
UPDATE cart_items
SET quantity = $3, updated_at = now()
WHERE cart_id = $1 AND id = $2
RETURNING id, cart_id, line_type, product_id, combo_id, quantity, created_at, updated_at;

-- name: DeleteCartItem :execrows
DELETE FROM cart_items
WHERE cart_id = $1 AND id = $2;

-- name: DeleteAllCartItems :exec
DELETE FROM cart_items
WHERE cart_id = $1;

-- name: CountCartItems :one
SELECT COUNT(*)::bigint AS count
FROM cart_items
WHERE cart_id = $1;
