-- name: CreateAuditLog :exec
INSERT INTO audit_logs (
    actor_id,
    actor_role,
    action,
    target_id,
    target_type,
    before_jsonb,
    after_jsonb,
    ip,
    user_agent
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, '')::inet, $9);

-- name: CountAuditLogsByTargetID :one
SELECT COUNT(*)::bigint AS total
FROM audit_logs
WHERE target_id = $1;

-- name: ListAuditLogsByTargetID :many
SELECT
    al.id,
    al.actor_id,
    al.actor_role,
    u.email AS actor_email,
    al.action,
    al.before_jsonb,
    al.after_jsonb,
    al.created_at
FROM audit_logs al
INNER JOIN users u ON u.id = al.actor_id
WHERE al.target_id = $1
ORDER BY al.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditLogsForSubjectBefore :many
-- What was recorded about one account — the account itself and its profile —
-- newest first, a page at a time by keyset, as the account holder's data
-- export reads it.
SELECT
    id,
    actor_id,
    actor_role,
    action,
    before_jsonb,
    after_jsonb,
    COALESCE(host(ip), '')::text AS ip,
    user_agent,
    created_at
FROM audit_logs
WHERE target_type IN ('user', 'user_profile')
  AND target_id = sqlc.arg('subject_id')
  AND (
    sqlc.narg('before_at')::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg('before_at')::timestamptz, sqlc.narg('before_id')::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit');

-- name: ScrubAuditLogsForSubject :exec
-- Removes an erased account's personal data from the audit trail and keeps
-- the trail: changes to the account and its profile lose their before and
-- after values (a profile change holds a name or phone number), and the
-- account's own actions lose the network and browser they came from. What
-- happened, when, and by which role stays on record.
UPDATE audit_logs
SET before_jsonb = CASE
      WHEN target_type IN ('user', 'user_profile') AND target_id = sqlc.arg('subject_id') THEN '{}'::jsonb
      ELSE before_jsonb
    END,
    after_jsonb = CASE
      WHEN target_type IN ('user', 'user_profile') AND target_id = sqlc.arg('subject_id') THEN '{}'::jsonb
      ELSE after_jsonb
    END,
    ip = CASE WHEN actor_id = sqlc.arg('subject_id') THEN NULL ELSE ip END,
    user_agent = CASE WHEN actor_id = sqlc.arg('subject_id') THEN NULL ELSE user_agent END
WHERE actor_id = sqlc.arg('subject_id')
   OR (target_type IN ('user', 'user_profile') AND target_id = sqlc.arg('subject_id'));
