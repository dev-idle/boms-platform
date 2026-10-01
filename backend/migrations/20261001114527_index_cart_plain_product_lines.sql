-- atlas:txmode none

-- Create index "cart_items_cart_plain_product_idx" to table: "cart_items", before the
-- old one goes, so plain product lines are never without a unique index.
CREATE UNIQUE INDEX CONCURRENTLY "cart_items_cart_plain_product_idx" ON "cart_items" ("cart_id", "product_id") WHERE ((line_type = 'product'::line_type) AND (configuration = '{}'::jsonb));
