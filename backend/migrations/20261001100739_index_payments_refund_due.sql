-- atlas:txmode none

-- Create index "payments_refund_due_idx" to table: "payments"
CREATE INDEX CONCURRENTLY "payments_refund_due_idx" ON "payments" ("refund_requested_at") WHERE ((status = 'captured'::payment_status) AND (refund_requested_at IS NOT NULL));
