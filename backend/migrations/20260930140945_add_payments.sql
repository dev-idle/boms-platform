-- Add value to enum type: "order_status"
ALTER TYPE "order_status" ADD VALUE 'awaiting_payment';
-- Create enum type "payment_provider"
CREATE TYPE "payment_provider" AS ENUM ('paypal');
-- Create enum type "payment_status"
CREATE TYPE "payment_status" AS ENUM ('created', 'pending', 'captured', 'denied');
-- Modify "discount_codes" table
ALTER TABLE "discount_codes" ADD CONSTRAINT "discount_codes_max_uses_per_customer_check" CHECK ((max_uses_per_customer IS NULL) OR (max_uses_per_customer > 0)), ADD COLUMN "max_uses_per_customer" integer NULL;
-- Modify "orders" table
ALTER TABLE "orders" ADD COLUMN "checkout_key" uuid NULL;
-- Create "payments" table
CREATE TABLE "payments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "provider" "payment_provider" NOT NULL,
  "provider_order_id" text NOT NULL,
  "approve_url" text NOT NULL,
  "status" "payment_status" NOT NULL DEFAULT 'created',
  "capture_id" text NULL,
  "amount_cents" bigint NOT NULL,
  "currency" text NOT NULL,
  "captured_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "payments_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "payments_amount_cents_check" CHECK (amount_cents > 0),
  CONSTRAINT "payments_capture_check" CHECK ((status = 'created'::payment_status) = (capture_id IS NULL)),
  CONSTRAINT "payments_captured_at_check" CHECK ((status = 'captured'::payment_status) = (captured_at IS NOT NULL)),
  CONSTRAINT "payments_currency_check" CHECK (currency ~ '^[A-Z]{3}$'::text)
);
-- Create index "payments_capture_idx" to table: "payments"
CREATE UNIQUE INDEX "payments_capture_idx" ON "payments" ("provider", "capture_id") WHERE (capture_id IS NOT NULL);
-- Create index "payments_order_id_idx" to table: "payments"
CREATE UNIQUE INDEX "payments_order_id_idx" ON "payments" ("order_id");
-- Create index "payments_provider_order_idx" to table: "payments"
CREATE UNIQUE INDEX "payments_provider_order_idx" ON "payments" ("provider", "provider_order_id");
