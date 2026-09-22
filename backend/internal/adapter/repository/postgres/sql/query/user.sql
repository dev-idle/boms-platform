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

-- name: AdminGetByID :one
SELECT id, email, password_hash, role, email_verified_at, must_change_password, created_at, updated_at, deleted_at
FROM users
WHERE id = $1;

-- name: AdminGetByIDForUpdate :one
-- Disabled rows included, like AdminGetByID: the caller reports "user is disabled".
SELECT id, email, password_hash, role, email_verified_at, must_change_password, created_at, updated_at, deleted_at
FROM users
WHERE id = $1
FOR UPDATE;

-- name: AdminRestore :execrows
UPDATE users
SET deleted_at = NULL,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NOT NULL;

-- name: AdminUpdateUserPassword :execrows
UPDATE users
SET password_hash = $2,
    must_change_password = true,
    updated_at = now()
WHERE id = $1;

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
    COALESCE(sp.employee_code, ''::citext) AS employee_code
FROM users u
LEFT JOIN customer_profiles cp ON cp.user_id = u.id
LEFT JOIN staff_profiles sp ON sp.user_id = u.id
LEFT JOIN admin_profiles ap ON ap.user_id = u.id
WHERE (
    sqlc.arg('search')::text = ''
    OR u.email ILIKE '%' || sqlc.arg('search') || '%'
    OR COALESCE(cp.display_name, sp.full_name, ap.full_name, '') ILIKE '%' || sqlc.arg('search') || '%'
    OR COALESCE(sp.employee_code::text, '') ILIKE '%' || sqlc.arg('search') || '%'
  )
  AND (
    sqlc.arg('role_filter')::text = ''
    OR u.role::text = sqlc.arg('role_filter')
  )
ORDER BY u.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: AdminListCount :one
SELECT COUNT(*)
FROM users u
LEFT JOIN customer_profiles cp ON cp.user_id = u.id
LEFT JOIN staff_profiles sp ON sp.user_id = u.id
LEFT JOIN admin_profiles ap ON ap.user_id = u.id
WHERE (
    sqlc.arg('search')::text = ''
    OR u.email ILIKE '%' || sqlc.arg('search') || '%'
    OR COALESCE(cp.display_name, sp.full_name, ap.full_name, '') ILIKE '%' || sqlc.arg('search') || '%'
    OR COALESCE(sp.employee_code::text, '') ILIKE '%' || sqlc.arg('search') || '%'
  )
  AND (
    sqlc.arg('role_filter')::text = ''
    OR u.role::text = sqlc.arg('role_filter')
  );

-- name: LockPhone :exec
-- Phones live in three profile tables, so no single unique index can guard them.
-- This lock (namespace 8734212, keyed by the number) serializes writers of one
-- number until the transaction ends.
SELECT pg_advisory_xact_lock(8734212, hashtext(sqlc.arg('phone')::text));

-- name: PhoneHeldByOtherActiveUser :one
SELECT EXISTS (
  SELECT 1
  FROM (
    SELECT user_id FROM customer_profiles WHERE phone = sqlc.arg('phone')::text
    UNION ALL
    SELECT user_id FROM staff_profiles WHERE phone = sqlc.arg('phone')::text
    UNION ALL
    SELECT user_id FROM admin_profiles WHERE phone = sqlc.arg('phone')::text
  ) AS holder
  JOIN users u ON u.id = holder.user_id
  WHERE u.id <> sqlc.arg('user_id')::uuid
    AND u.deleted_at IS NULL
) AS held;

-- name: ReleasePhone :exec
-- Clears the phone on whichever profile the user has. Used when a returning
-- account finds its number taken by an active one: the active holder keeps it.
WITH released_customer AS (
  UPDATE customer_profiles SET phone = NULL, updated_at = now() WHERE user_id = sqlc.arg('user_id')::uuid
), released_staff AS (
  UPDATE staff_profiles SET phone = NULL, updated_at = now() WHERE user_id = sqlc.arg('user_id')::uuid
)
UPDATE admin_profiles SET phone = NULL, updated_at = now() WHERE user_id = sqlc.arg('user_id')::uuid;
