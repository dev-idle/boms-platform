-- name: PublishPolicyVersion :one
-- Records the policy version an API serves and returns when one first served
-- it: a version served before keeps its time.
INSERT INTO policy_versions (version)
VALUES ($1)
ON CONFLICT (version) DO UPDATE SET version = EXCLUDED.version
RETURNING published_at;
