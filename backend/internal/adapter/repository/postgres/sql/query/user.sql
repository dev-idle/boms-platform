-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, email_verified_at, created_at, updated_at, deleted_at
FROM users
WHERE email = $1
  AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, password_hash, role, email_verified_at, created_at, updated_at, deleted_at
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id, email, password_hash, role, email_verified_at, created_at, updated_at, deleted_at;

-- name: UpdateUserPassword :execrows
UPDATE users
SET password_hash = $2,
    updated_at    = now()
WHERE id = $1
  AND deleted_at IS NULL;

-- name: SoftDeleteUser :execrows
UPDATE users
SET deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL;
