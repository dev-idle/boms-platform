-- name: CreateCustomerProfile :one
INSERT INTO customer_profiles (user_id, display_name, phone)
VALUES ($1, $2, $3)
RETURNING user_id, display_name, phone, created_at, updated_at;

-- name: GetCustomerProfileByUserID :one
SELECT user_id, display_name, phone, created_at, updated_at
FROM customer_profiles
WHERE user_id = $1;

-- name: UpdateCustomerProfileByUserID :one
UPDATE customer_profiles
SET display_name = $2,
    phone = $3,
    updated_at = now()
WHERE user_id = $1
RETURNING user_id, display_name, phone, created_at, updated_at;

-- name: DeleteCustomerProfileByUserID :execrows
DELETE FROM customer_profiles
WHERE user_id = $1;
