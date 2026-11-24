-- atlas:txmode none

-- Create index "conversations_customer_general_idx" to table: "conversations"
CREATE UNIQUE INDEX CONCURRENTLY "conversations_customer_general_idx" ON "conversations" ("customer_id") WHERE (order_id IS NULL);
