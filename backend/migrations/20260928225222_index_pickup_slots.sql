-- atlas:txmode none

-- orders is live: build the index without blocking checkouts.
-- Create index "orders_pickup_slot_idx" to table: "orders"
CREATE INDEX CONCURRENTLY "orders_pickup_slot_idx" ON "orders" ("pickup_at") WHERE (status <> 'cancelled'::order_status);
