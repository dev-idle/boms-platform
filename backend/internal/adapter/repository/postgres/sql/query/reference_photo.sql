-- name: RequestReferencePhotoErasure :exec
-- Asks, with the erasure, for the customer's reference photos to be deleted.
INSERT INTO reference_photo_erasures (user_id)
VALUES ($1)
ON CONFLICT (user_id) DO NOTHING;

-- name: ListReferencePhotoErasuresDue :many
-- The customers whose reference photos are still to be deleted, the oldest
-- request first.
SELECT user_id
FROM reference_photo_erasures
WHERE erased_at IS NULL
ORDER BY requested_at
LIMIT $1;

-- name: MarkReferencePhotosErased :exec
UPDATE reference_photo_erasures
SET erased_at = now()
WHERE user_id = $1
  AND erased_at IS NULL;
