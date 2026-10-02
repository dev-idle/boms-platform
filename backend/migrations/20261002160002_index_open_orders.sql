-- atlas:txmode none

-- Create index "orders_open_pickup_idx" to table: "orders", before the
-- production one it replaces goes, so the kitchen's queue keeps an index.
CREATE INDEX CONCURRENTLY "orders_open_pickup_idx" ON "orders" ("status", "pickup_at") WHERE (status = ANY (ARRAY['pending'::order_status, 'confirmed'::order_status, 'in_production'::order_status, 'ready'::order_status]));
