-- atlas:txmode none

-- Create index "order_status_events_ready_created_idx" to table: "order_status_events"
CREATE INDEX CONCURRENTLY "order_status_events_ready_created_idx" ON "order_status_events" ("created_at") WHERE (to_status = 'ready'::order_status);
