-- Add value to enum type: "order_status"
ALTER TYPE "order_status" ADD VALUE 'expired';
-- Modify "order_status_events" table
ALTER TABLE "order_status_events" ADD CONSTRAINT "order_status_events_actor_check" CHECK ((actor_id IS NULL) = (actor_role IS NULL)), ALTER COLUMN "actor_id" DROP NOT NULL, ALTER COLUMN "actor_role" DROP NOT NULL;
-- Modify "orders" table
ALTER TABLE "orders" ADD COLUMN "payment_due_at" timestamptz NULL;
-- Modify "store_settings" table
ALTER TABLE "store_settings" ADD CONSTRAINT "store_settings_payment_hold_check" CHECK ((payment_hold_minutes >= 5) AND (payment_hold_minutes <= 120)), ADD COLUMN "payment_hold_minutes" smallint NOT NULL DEFAULT 15;
