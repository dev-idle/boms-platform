-- Create enum type "order_channel"
CREATE TYPE "order_channel" AS ENUM ('online', 'counter', 'phone');
-- Modify "orders" table
ALTER TABLE "orders" ADD CONSTRAINT "orders_customer_check" CHECK (((user_id IS NULL) = (guest_name IS NOT NULL)) AND ((guest_name IS NULL) = (guest_phone IS NULL)) AND ((channel <> 'online'::order_channel) OR (user_id IS NOT NULL))), ALTER COLUMN "user_id" DROP NOT NULL, ADD COLUMN "channel" "order_channel" NOT NULL DEFAULT 'online', ADD COLUMN "guest_name" text NULL, ADD COLUMN "guest_phone" text NULL;
-- Modify "payments" table
ALTER TABLE "payments" DROP CONSTRAINT "payments_capture_check", ADD CONSTRAINT "payments_capture_check" CHECK ((provider <> 'paypal'::payment_provider) OR ((status = 'created'::payment_status) = (capture_id IS NULL))), ADD CONSTRAINT "payments_cash_check" CHECK ((provider <> 'cash'::payment_provider) OR ((capture_id IS NULL) AND (status = ANY (ARRAY['created'::payment_status, 'captured'::payment_status])) AND (refund_requested_at IS NULL))), ADD CONSTRAINT "payments_provider_check" CHECK ((provider = 'paypal'::payment_provider) = ((provider_order_id IS NOT NULL) AND (approve_url IS NOT NULL))), ALTER COLUMN "provider_order_id" DROP NOT NULL, ALTER COLUMN "approve_url" DROP NOT NULL;
-- Modify "products" table
ALTER TABLE "products" ADD COLUMN "sold_out_on" date NULL;
