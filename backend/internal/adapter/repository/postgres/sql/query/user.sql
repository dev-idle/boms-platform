-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, email_verified_at, must_change_password, created_at, updated_at, deleted_at
FROM users
WHERE email = $1
  AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, password_hash, role, email_verified_at, must_change_password, created_at, updated_at, deleted_at
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetUserByIDForUpdate :one
SELECT id, email, password_hash, role, email_verified_at, must_change_password, created_at, updated_at, deleted_at
FROM users
WHERE id = $1
  AND deleted_at IS NULL
FOR UPDATE;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, role, must_change_password)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, role, email_verified_at, must_change_password, created_at, updated_at, deleted_at;

-- name: AdminCreate :one
INSERT INTO users (email, password_hash, role, must_change_password)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, role, email_verified_at, must_change_password, created_at, updated_at, deleted_at;

-- name: UpdateUserPassword :execrows
UPDATE users
SET password_hash = $2,
    updated_at    = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpdateRole :execrows
UPDATE users
SET role = $2,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: SetMustChangePassword :execrows
UPDATE users
SET must_change_password = true,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: ClearMustChangePassword :execrows
UPDATE users
SET must_change_password = false,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: SoftDelete :execrows
UPDATE users
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: AdminList :many
SELECT
    u.id,
    u.email,
    u.role,
    u.email_verified_at,
    u.must_change_password,
    u.created_at,
    u.updated_at,
    u.deleted_at,
    cp.display_name,
    COALESCE(sp.full_name, ap.full_name, '') AS full_name,
    COALESCE(sp.phone, ap.phone, cp.phone) AS phone,
    sp.employee_code
FROM users u
LEFT JOIN customer_profiles cp ON cp.user_id = u.id
LEFT JOIN staff_profiles sp ON sp.user_id = u.id
LEFT JOIN admin_profiles ap ON ap.user_id = u.id
WHERE u.deleted_at IS NULL
  AND (
    $1::text = ''
    OR u.email ILIKE '%' || $1 || '%'
    OR COALESCE(cp.display_name, sp.full_name, ap.full_name, '') ILIKE '%' || $1 || '%'
    OR COALESCE(sp.employee_code::text, '') ILIKE '%' || $1 || '%'
  )
ORDER BY u.created_at DESC
LIMIT $2 OFFSET $3;

-- name: AdminListCount :one
SELECT COUNT(*)
FROM users u
LEFT JOIN customer_profiles cp ON cp.user_id = u.id
LEFT JOIN staff_profiles sp ON sp.user_id = u.id
LEFT JOIN admin_profiles ap ON ap.user_id = u.id
WHERE u.deleted_at IS NULL
  AND (
    $1::text = ''
    OR u.email ILIKE '%' || $1 || '%'
    OR COALESCE(cp.display_name, sp.full_name, ap.full_name, '') ILIKE '%' || $1 || '%'
    OR COALESCE(sp.employee_code::text, '') ILIKE '%' || $1 || '%'
  );
