-- Add value to enum type: "order_incident_type"
ALTER TYPE "order_incident_type" ADD VALUE 'refund_failed';
-- Add value to enum type: "order_incident_type"
ALTER TYPE "order_incident_type" ADD VALUE 'payment_reversed';
-- Modify "payments" table
ALTER TABLE "payments" ADD CONSTRAINT "payments_refund_attempts_check" CHECK ((refund_attempts >= 0) AND ((refund_last_error IS NULL) OR (char_length(refund_last_error) <= 500))), ADD COLUMN "refund_attempts" integer NOT NULL DEFAULT 0, ADD COLUMN "refund_last_error" text NULL, ADD COLUMN "refund_retry_at" timestamptz NULL;
-- Create "payment_refund_notices" table
CREATE TABLE "payment_refund_notices" (
  "refund_id" text NOT NULL,
  "payment_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("refund_id"),
  CONSTRAINT "payment_refund_notices_payment_id_fkey" FOREIGN KEY ("payment_id") REFERENCES "payments" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
