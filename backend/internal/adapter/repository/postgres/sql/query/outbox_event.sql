-- Every time below comes from the database clock (created_at defaults to
-- clock_timestamp()), so the API writing rows and the worker sweeping them never
-- compare two hosts' clocks.

-- name: InsertOutboxEvent :one
INSERT INTO outbox_events (id, topic, audience, data)
VALUES ($1, $2, $3, $4)
RETURNING created_at;

-- name: ClaimUnpublishedOutboxEvents :many
-- Rows another sweeper already holds are skipped, not waited on, so two workers
-- never claim the same row at once.
SELECT id, topic, audience, data, created_at
FROM outbox_events
WHERE published_at IS NULL
  AND created_at < clock_timestamp() - make_interval(secs => sqlc.arg('grace_seconds')::double precision)
ORDER BY created_at ASC
LIMIT sqlc.arg('limit')
FOR UPDATE SKIP LOCKED;

-- name: MarkOutboxEventsPublished :exec
UPDATE outbox_events
SET published_at = clock_timestamp()
WHERE id = ANY(sqlc.arg('ids')::uuid[])
  AND published_at IS NULL;

-- name: RecordOutboxPublishFailure :exec
UPDATE outbox_events
SET attempts = attempts + 1,
    last_error = sqlc.arg('last_error')::text
WHERE id = ANY(sqlc.arg('ids')::uuid[])
  AND published_at IS NULL;

-- name: DeletePublishedOutboxEvents :execrows
DELETE FROM outbox_events
WHERE published_at < clock_timestamp() - make_interval(secs => sqlc.arg('retention_seconds')::double precision);
