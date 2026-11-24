-- name: ListRecentAssistantTurns :many
-- The last turns of a session that belongs to the visitor asking (no user
-- for one not signed in), newest first; none for a session of someone else.
-- A question and its answer share their transaction's time, and the question
-- sorts first by its role.
SELECT m.role, m.body, m.created_at
FROM assistant_messages m
JOIN assistant_sessions s ON s.id = m.session_id
WHERE m.session_id = sqlc.arg('session_id')
  AND s.user_id IS NOT DISTINCT FROM sqlc.narg('user_id')::uuid
ORDER BY m.created_at DESC, m.role DESC
LIMIT sqlc.arg('limit');

-- name: TouchAssistantSession :one
-- Starts the session, or marks the time of its new message. A session never
-- changes hands: one that belongs to someone else is left as it is, and no
-- row comes back.
INSERT INTO assistant_sessions (id, user_id)
VALUES (sqlc.arg('id'), sqlc.narg('user_id'))
ON CONFLICT (id) DO UPDATE
SET last_message_at = now()
WHERE assistant_sessions.user_id IS NOT DISTINCT FROM EXCLUDED.user_id
RETURNING id;

-- name: CreateAssistantTurn :exec
-- A question and its answer, both at the transaction's time.
INSERT INTO assistant_messages (session_id, role, body)
VALUES (sqlc.arg('session_id'), 'visitor', sqlc.arg('question')),
       (sqlc.arg('session_id'), 'assistant', sqlc.arg('answer'));

-- name: DeleteIdleAssistantSessions :execrows
-- Sessions with no message for longer than their retention, their turns with them.
DELETE FROM assistant_sessions
WHERE last_message_at < clock_timestamp() - make_interval(secs => sqlc.arg('retention_seconds')::double precision);

-- name: ListAssistantTurnsForUser :many
-- Every turn of the account's sessions, newest session first, each oldest first.
SELECT
  s.id AS session_id,
  s.created_at AS started_at,
  m.role,
  m.body,
  m.created_at
FROM assistant_sessions s
JOIN assistant_messages m ON m.session_id = s.id
WHERE s.user_id = $1
ORDER BY s.created_at DESC, s.id, m.created_at, m.role;

-- name: DeleteAssistantSessionsForUser :exec
DELETE FROM assistant_sessions
WHERE user_id = $1;
