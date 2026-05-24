-- name: CreateStaffProfile :one
INSERT INTO staff_profiles (user_id, full_name, phone, employee_code, hire_date, shift)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING user_id, full_name, phone, employee_code, hire_date, shift, created_at, updated_at;

-- name: GetStaffProfileByUserID :one
SELECT user_id, full_name, phone, employee_code, hire_date, shift, created_at, updated_at
FROM staff_profiles
WHERE user_id = $1;

-- name: UpdateStaffProfileByUserID :one
UPDATE staff_profiles
SET full_name = $2,
    phone = $3,
    employee_code = $4,
    hire_date = $5,
    shift = $6,
    updated_at = now()
WHERE user_id = $1
RETURNING user_id, full_name, phone, employee_code, hire_date, shift, created_at, updated_at;

-- name: DeleteStaffProfileByUserID :execrows
DELETE FROM staff_profiles
WHERE user_id = $1;
