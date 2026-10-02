-- Create enum type "promotion_status"
CREATE TYPE "promotion_status" AS ENUM ('sending', 'sent');
-- Modify "customer_profiles" table
ALTER TABLE "customer_profiles" ADD COLUMN "marketing_consent_at" timestamptz NULL;
-- Create "promotions" table
CREATE TABLE "promotions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "subject" text NOT NULL,
  "body" text NOT NULL,
  "status" "promotion_status" NOT NULL DEFAULT 'sending',
  "created_by" uuid NOT NULL,
  "recipient_count" integer NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "promotions_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "promotions_body_check" CHECK ((char_length(body) >= 1) AND (char_length(body) <= 5000)),
  CONSTRAINT "promotions_recipient_count_check" CHECK (recipient_count >= 0),
  CONSTRAINT "promotions_sent_check" CHECK ((status = 'sent'::promotion_status) = (recipient_count IS NOT NULL)),
  CONSTRAINT "promotions_subject_check" CHECK ((char_length(subject) >= 1) AND (char_length(subject) <= 120))
);
-- Create index "promotions_created_at_idx" to table: "promotions"
CREATE INDEX "promotions_created_at_idx" ON "promotions" ("created_at");
