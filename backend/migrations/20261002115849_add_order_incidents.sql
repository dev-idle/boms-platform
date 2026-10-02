-- Create enum type "order_incident_type"
CREATE TYPE "order_incident_type" AS ENUM ('bakery_cancelled', 'ready_late', 'no_show', 'payment_failed', 'payment_expired', 'refunded', 'payment_anomaly', 'wrong_items', 'custom_mismatch', 'other');
-- Modify "payments" table
ALTER TABLE "payments" DROP CONSTRAINT "payments_capture_check", ADD CONSTRAINT "payments_capture_check" CHECK ((provider <> 'paypal'::payment_provider) OR (status = 'denied'::payment_status) OR ((status = 'created'::payment_status) = (capture_id IS NULL)));
-- Create "order_incidents" table
CREATE TABLE "order_incidents" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "type" "order_incident_type" NOT NULL,
  "note" text NULL,
  "actor_id" uuid NULL,
  "actor_role" "user_role" NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_incidents_actor_id_fkey" FOREIGN KEY ("actor_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "order_incidents_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "order_incidents_actor_check" CHECK ((actor_id IS NULL) = (actor_role IS NULL)),
  CONSTRAINT "order_incidents_note_check" CHECK ((note IS NULL) OR (((char_length(note) >= 1) AND (char_length(note) <= 500)) AND (type = ANY (ARRAY['wrong_items'::order_incident_type, 'custom_mismatch'::order_incident_type, 'other'::order_incident_type])))),
  CONSTRAINT "order_incidents_reported_check" CHECK ((type <> ALL (ARRAY['wrong_items'::order_incident_type, 'custom_mismatch'::order_incident_type, 'other'::order_incident_type])) OR (actor_id IS NOT NULL))
);
-- Create index "order_incidents_created_at_idx" to table: "order_incidents"
CREATE INDEX "order_incidents_created_at_idx" ON "order_incidents" ("created_at");
-- Create index "order_incidents_order_created_idx" to table: "order_incidents"
CREATE INDEX "order_incidents_order_created_idx" ON "order_incidents" ("order_id", "created_at");
