-- name: CreateStaffProfile :one
INSERT INTO staff_profiles (user_id, full_name, phone, employee_code)
VALUES ($1, $2, $3, $4)
RETURNING user_id, full_name, phone, employee_code, created_at, updated_at;

-- name: GetStaffProfileByUserID :one
SELECT user_id, full_name, phone, employee_code, created_at, updated_at
FROM staff_profiles
WHERE user_id = $1;

-- name: UpdateStaffProfileByUserID :one
UPDATE staff_profiles
SET full_name = $2,
    phone = $3,
    employee_code = $4,
    updated_at = now()
WHERE user_id = $1
RETURNING user_id, full_name, phone, employee_code, created_at, updated_at;

-- name: DeleteStaffProfileByUserID :execrows
DELETE FROM staff_profiles
WHERE user_id = $1;

-- name: NextEmployeeCode :one
SELECT
  (
    'EMP-' || lpad(
      (
        SELECT COALESCE(
          MAX(
            CASE
              WHEN employee_code::text ~ '^EMP-[0-9]+$' THEN
                NULLIF(regexp_replace(employee_code::text, '^EMP-0*', ''), '')::bigint
            END
          ),
          0
        ) + 1
        FROM staff_profiles
      )::text,
      5,
      '0'
    )
  )::text AS next_code
FROM (SELECT pg_advisory_xact_lock(8734211)) AS lock;
