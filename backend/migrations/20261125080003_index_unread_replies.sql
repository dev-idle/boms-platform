-- atlas:txmode none

-- Create index "conversations_unread_replies_idx" to table: "conversations"
CREATE INDEX CONCURRENTLY "conversations_unread_replies_idx" ON "conversations" ("last_message_at") WHERE ((unread_by_customer > 0) AND (reply_emailed_at IS NULL));
