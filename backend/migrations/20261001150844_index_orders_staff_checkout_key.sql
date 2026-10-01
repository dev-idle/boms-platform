-- atlas:txmode none

-- Create index "orders_staff_checkout_key_idx" to table: "orders"
CREATE UNIQUE INDEX CONCURRENTLY "orders_staff_checkout_key_idx" ON "orders" ("checkout_key") WHERE ((channel <> 'online'::order_channel) AND (checkout_key IS NOT NULL));
