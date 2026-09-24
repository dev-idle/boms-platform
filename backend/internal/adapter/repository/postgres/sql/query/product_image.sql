-- name: DeleteProductImagesByProductID :exec
DELETE FROM product_images
WHERE product_id = $1;

-- name: InsertProductImage :one
INSERT INTO product_images (product_id, sort_order, image_url)
VALUES ($1, $2, $3)
RETURNING id, product_id, sort_order, image_url, created_at, updated_at;

-- name: ListProductImagesByProductID :many
SELECT image_url, sort_order
FROM product_images
WHERE product_id = $1
ORDER BY sort_order ASC;
