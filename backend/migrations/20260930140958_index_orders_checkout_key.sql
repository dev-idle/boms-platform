-- atlas:txmode none

-- orders is live: build the index without blocking checkouts.
-- Create index "orders_checkout_key_idx" to table: "orders"
CREATE UNIQUE INDEX CONCURRENTLY "orders_checkout_key_idx" ON "orders" ("user_id", "checkout_key") WHERE (checkout_key IS NOT NULL);
