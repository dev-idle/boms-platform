-- name: ReplaceUserToken :exec
-- Issuing a token replaces the user's last one of that purpose, so only the
-- newest link works. Expiry is set on the database clock, which redemption
-- checks against.
INSERT INTO user_tokens (user_id, purpose, token_hash, expires_at)
VALUES (
  sqlc.arg('user_id'),
  sqlc.arg('purpose'),
  sqlc.arg('token_hash'),
  now() + make_interval(secs => sqlc.arg('ttl_seconds')::double precision)
)
ON CONFLICT (user_id, purpose) DO UPDATE
SET token_hash = EXCLUDED.token_hash,
    expires_at = EXCLUDED.expires_at,
    created_at = now();

-- name: RedeemUserToken :one
-- Single use: the token goes as it is redeemed, in one statement, so two
-- requests with the same link cannot both succeed. An expired token, or one
-- of a closed account, opens nothing.
DELETE FROM user_tokens t
USING users u
WHERE t.token_hash = sqlc.arg('token_hash')
  AND t.purpose = sqlc.arg('purpose')
  AND t.expires_at > now()
  AND u.id = t.user_id
  AND u.deleted_at IS NULL
RETURNING t.user_id;

-- name: DeleteUserTokens :exec
DELETE FROM user_tokens
WHERE user_id = $1;
