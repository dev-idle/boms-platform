-- Create enum type "station"
CREATE TYPE "station" AS ENUM ('kitchen', 'counter');
-- Create enum type "order_type"
CREATE TYPE "order_type" AS ENUM ('instant', 'pre_order');
-- Modify "categories" table
ALTER TABLE "categories" ADD COLUMN "station" "station" NOT NULL DEFAULT 'kitchen';
-- Modify "orders" table. Every category starts in the kitchen, so every order
-- placed so far was a pre-order; new orders must say which they are.
ALTER TABLE "orders" ADD COLUMN "order_type" "order_type" NOT NULL DEFAULT 'pre_order';
ALTER TABLE "orders" ALTER COLUMN "order_type" DROP DEFAULT;
-- Modify "products" table
ALTER TABLE "products" ADD CONSTRAINT "products_lead_time_minutes_check" CHECK ((lead_time_minutes >= 0) AND (lead_time_minutes <= 10080)), ADD COLUMN "lead_time_minutes" integer NOT NULL DEFAULT 0;
-- Modify "store_settings" table
ALTER TABLE "store_settings" ADD CONSTRAINT "store_settings_capacity_check" CHECK ((slot_capacity >= 1) AND (slot_capacity <= 200)), ADD CONSTRAINT "store_settings_instant_prep_check" CHECK ((instant_prep_minutes >= 0) AND (instant_prep_minutes <= 240)), ADD CONSTRAINT "store_settings_slot_check" CHECK ((slot_minutes = ANY (ARRAY[10, 15, 20, 30, 60])) AND ((closes_at_minute - opens_at_minute) >= slot_minutes)), ADD COLUMN "slot_minutes" smallint NOT NULL DEFAULT 30, ADD COLUMN "slot_capacity" smallint NOT NULL DEFAULT 10, ADD COLUMN "instant_prep_minutes" smallint NOT NULL DEFAULT 20;
