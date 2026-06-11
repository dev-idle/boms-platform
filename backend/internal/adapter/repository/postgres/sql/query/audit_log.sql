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
