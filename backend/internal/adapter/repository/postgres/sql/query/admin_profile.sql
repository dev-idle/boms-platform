-- name: CreateAdminProfile :one
INSERT INTO admin_profiles (user_id, full_name, phone)
VALUES ($1, $2, $3)
RETURNING user_id, full_name, phone, created_at, updated_at;

-- name: GetAdminProfileByUserID :one
SELECT user_id, full_name, phone, created_at, updated_at
FROM admin_profiles
WHERE user_id = $1;

-- name: UpdateAdminProfileByUserID :one
UPDATE admin_profiles
SET full_name = $2,
    phone = $3,
    updated_at = now()
WHERE user_id = $1
RETURNING user_id, full_name, phone, created_at, updated_at;

-- name: DeleteAdminProfileByUserID :execrows
DELETE FROM admin_profiles
WHERE user_id = $1;
