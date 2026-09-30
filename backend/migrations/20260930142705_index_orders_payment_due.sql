-- atlas:txmode none

-- orders is live: build the index without blocking checkouts.
-- Create index "orders_payment_due_idx" to table: "orders"
CREATE INDEX CONCURRENTLY "orders_payment_due_idx" ON "orders" ("payment_due_at") WHERE (status = 'awaiting_payment'::order_status);
