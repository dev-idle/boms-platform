-- name: GetStoreSettings :one
SELECT opens_at_minute, closes_at_minute, preorder_min_lead_minutes, max_advance_days, updated_at
FROM store_settings
WHERE id = 1;

-- name: GetStoreSettingsForUpdate :one
-- Locks the row so two admins editing at once cannot undo each other.
SELECT opens_at_minute, closes_at_minute, preorder_min_lead_minutes, max_advance_days, updated_at
FROM store_settings
WHERE id = 1
FOR UPDATE;

-- name: UpdateStoreSettings :one
UPDATE store_settings
SET opens_at_minute = sqlc.arg(opens_at_minute),
    closes_at_minute = sqlc.arg(closes_at_minute),
    preorder_min_lead_minutes = sqlc.arg(preorder_min_lead_minutes),
    max_advance_days = sqlc.arg(max_advance_days),
    updated_at = now()
WHERE id = 1
RETURNING opens_at_minute, closes_at_minute, preorder_min_lead_minutes, max_advance_days, updated_at;

-- name: ListStoreClosedDates :many
-- Days in [from_day, to_day]; the caller bounds the window, and one row per
-- day keeps the result bounded by it.
SELECT id, closed_on, reason, created_at
FROM store_closed_dates
WHERE deleted_at IS NULL
  AND closed_on BETWEEN sqlc.arg(from_day)::date AND sqlc.arg(to_day)::date
ORDER BY closed_on
LIMIT sqlc.arg(max_rows)::int;

-- name: CreateStoreClosedDate :one
INSERT INTO store_closed_dates (closed_on, reason)
VALUES (sqlc.arg(closed_on), sqlc.arg(reason))
RETURNING id, closed_on, reason, created_at;

-- name: SoftDeleteStoreClosedDate :one
UPDATE store_closed_dates
SET deleted_at = now(), updated_at = now()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING id, closed_on, reason, created_at;
