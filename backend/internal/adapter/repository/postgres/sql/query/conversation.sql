-- name: OpenConversationForCustomerMessage :one
-- Records that the customer wrote about the order: the conversation starts,
-- or opens again, the counter has one more message to read, and the customer
-- has read everything before their own.
INSERT INTO conversations (order_id, unread_by_staff, last_message_at)
VALUES (sqlc.arg('order_id'), 1, now())
ON CONFLICT (order_id) DO UPDATE
SET status = 'open',
    unread_by_staff = conversations.unread_by_staff + 1,
    unread_by_customer = 0,
    last_message_at = GREATEST(conversations.last_message_at, now()),
    updated_at = now()
RETURNING id;

-- name: OpenConversationForStaffMessage :one
-- Records that staff wrote about the order: the conversation starts, or opens
-- again, and passes to the writer; the customer has one more message to read,
-- and the counter has read everything before it.
INSERT INTO conversations (order_id, assigned_staff_id, unread_by_customer, last_message_at)
VALUES (sqlc.arg('order_id'), sqlc.arg('staff_id'), 1, now())
ON CONFLICT (order_id) DO UPDATE
SET status = 'open',
    assigned_staff_id = EXCLUDED.assigned_staff_id,
    unread_by_customer = conversations.unread_by_customer + 1,
    unread_by_staff = 0,
    last_message_at = GREATEST(conversations.last_message_at, now()),
    updated_at = now()
RETURNING id;

-- name: CreateMessage :one
-- Writes a message and returns it as a thread lists it.
WITH created AS (
  INSERT INTO messages (conversation_id, author_id, author_role, body)
  VALUES (sqlc.arg('conversation_id'), sqlc.arg('author_id'), sqlc.arg('author_role'), sqlc.arg('body'))
  RETURNING id, author_id, author_role, body, created_at
)
SELECT
  created.id,
  created.author_role,
  created.body,
  created.created_at,
  sp.full_name AS author_name
FROM created
LEFT JOIN staff_profiles sp ON sp.user_id = created.author_id AND created.author_role = 'staff'::user_role;

-- name: GetConversationByOrderID :one
SELECT
  c.status,
  c.unread_by_customer,
  c.unread_by_staff,
  sp.full_name AS assigned_staff_name
FROM conversations c
LEFT JOIN staff_profiles sp ON sp.user_id = c.assigned_staff_id
WHERE c.order_id = $1;

-- name: ListMessagesByOrderID :many
-- A page of an order's messages, newest first: the latest, or those before
-- the message before_id names. A staff message carries its writer's name.
SELECT
  m.id,
  m.author_role,
  m.body,
  m.created_at,
  sp.full_name AS author_name
FROM messages m
JOIN conversations c ON c.id = m.conversation_id
LEFT JOIN staff_profiles sp ON sp.user_id = m.author_id AND m.author_role = 'staff'::user_role
WHERE c.order_id = sqlc.arg('order_id')
  AND m.deleted_at IS NULL
  AND (
    sqlc.narg('before_id')::uuid IS NULL
    OR (m.created_at, m.id) < (
      SELECT b.created_at, b.id FROM messages b
      WHERE b.id = sqlc.narg('before_id')::uuid AND b.conversation_id = c.id
    )
  )
ORDER BY m.created_at DESC, m.id DESC
LIMIT sqlc.arg('limit');

-- name: MarkConversationReadByCustomer :execrows
UPDATE conversations
SET unread_by_customer = 0,
    updated_at = now()
WHERE order_id = $1
  AND unread_by_customer > 0;

-- name: MarkConversationReadByStaff :execrows
UPDATE conversations
SET unread_by_staff = 0,
    updated_at = now()
WHERE order_id = $1
  AND unread_by_staff > 0;

-- name: SetConversationStatus :execrows
-- Resolving a conversation reads it for the counter: nothing in it waits on them.
UPDATE conversations
SET status = sqlc.arg('status'),
    unread_by_staff = CASE WHEN sqlc.arg('status') = 'closed'::conversation_status THEN 0 ELSE unread_by_staff END,
    updated_at = now()
WHERE order_id = sqlc.arg('order_id')
  AND status <> sqlc.arg('status');

-- name: StaffListConversations :many
-- The counter's inbox, latest message first: each conversation with its
-- order, its customer and a preview of its last message. A customer whose
-- account is closed is not shown.
SELECT
  c.order_id,
  o.code AS order_code,
  c.status,
  c.unread_by_staff,
  c.last_message_at,
  cp.display_name AS customer_name,
  u.email AS customer_email,
  left(last.body, 200)::text AS last_body
FROM conversations c
JOIN orders o ON o.id = c.order_id
JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
LEFT JOIN customer_profiles cp ON cp.user_id = u.id
JOIN LATERAL (
  SELECT m.body
  FROM messages m
  WHERE m.conversation_id = c.id AND m.deleted_at IS NULL
  ORDER BY m.created_at DESC, m.id DESC
  LIMIT 1
) last ON true
WHERE (
    sqlc.narg('status')::conversation_status IS NULL
    OR c.status = sqlc.narg('status')::conversation_status
  )
ORDER BY c.last_message_at DESC, c.id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: StaffListConversationsCount :one
SELECT count(*)::bigint
FROM conversations c
JOIN orders o ON o.id = c.order_id
JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL
WHERE (
    sqlc.narg('status')::conversation_status IS NULL
    OR c.status = sqlc.narg('status')::conversation_status
  );

-- name: StaffConversationCounts :one
-- How many conversations wait on the counter: open ones, and those holding
-- messages nobody at the counter has read.
SELECT
  count(*) FILTER (WHERE c.status = 'open'::conversation_status)::bigint AS open_count,
  count(*) FILTER (WHERE c.unread_by_staff > 0)::bigint AS unread_count
FROM conversations c
JOIN orders o ON o.id = c.order_id
JOIN users u ON u.id = o.user_id AND u.deleted_at IS NULL;

-- name: ListUnreadByCustomer :many
-- What the customer has not read yet on each of the orders given.
SELECT order_id, unread_by_customer
FROM conversations
WHERE order_id = ANY(sqlc.arg('order_ids')::uuid[])
  AND unread_by_customer > 0;

-- name: ListMessagesByOrderIDs :many
-- Every message on the orders given, oldest first, for a personal data export.
SELECT
  c.order_id,
  m.id,
  m.author_role,
  m.body,
  m.created_at
FROM messages m
JOIN conversations c ON c.id = m.conversation_id
WHERE c.order_id = ANY(sqlc.arg('order_ids')::uuid[])
  AND m.deleted_at IS NULL
ORDER BY c.order_id, m.created_at, m.id;

-- name: EraseCustomerMessages :exec
-- Erases the text of every message on the customer's orders, theirs and the
-- counter's: what was said is about them.
UPDATE messages m
SET body = '',
    deleted_at = now()
FROM conversations c
JOIN orders o ON o.id = c.order_id
WHERE m.conversation_id = c.id
  AND o.user_id = sqlc.arg('user_id')::uuid
  AND m.deleted_at IS NULL;

-- name: CloseCustomerConversations :exec
-- Closes the customer's conversations as their account is erased: nothing in
-- them is left to read.
UPDATE conversations c
SET status = 'closed',
    unread_by_customer = 0,
    unread_by_staff = 0,
    updated_at = now()
FROM orders o
WHERE o.id = c.order_id
  AND o.user_id = sqlc.arg('user_id')::uuid;
