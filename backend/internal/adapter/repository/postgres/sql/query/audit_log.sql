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
