-- Create "store_closed_dates" table
CREATE TABLE "store_closed_dates" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "closed_on" date NOT NULL,
  "reason" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "store_closed_dates_reason_check" CHECK ((char_length(reason) >= 1) AND (char_length(reason) <= 200))
);
-- Create index "store_closed_dates_day_idx" to table: "store_closed_dates"
CREATE UNIQUE INDEX "store_closed_dates_day_idx" ON "store_closed_dates" ("closed_on") WHERE (deleted_at IS NULL);
-- Create "store_settings" table
CREATE TABLE "store_settings" (
  "id" smallint NOT NULL DEFAULT 1,
  "opens_at_minute" smallint NOT NULL,
  "closes_at_minute" smallint NOT NULL,
  "preorder_min_lead_minutes" integer NOT NULL,
  "max_advance_days" smallint NOT NULL,
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "store_settings_advance_check" CHECK ((max_advance_days >= 1) AND (max_advance_days <= 90)),
  CONSTRAINT "store_settings_hours_check" CHECK ((opens_at_minute >= 0) AND (closes_at_minute < 1440) AND (opens_at_minute < closes_at_minute)),
  CONSTRAINT "store_settings_lead_check" CHECK ((preorder_min_lead_minutes >= 0) AND (preorder_min_lead_minutes <= 10080)),
  CONSTRAINT "store_settings_singleton_check" CHECK (id = 1)
);
-- Seed the single settings row with the pickup rules that were constants
-- before: 08:00-18:00, two hours ahead, up to 14 days ahead.
INSERT INTO "store_settings" ("id", "opens_at_minute", "closes_at_minute", "preorder_min_lead_minutes", "max_advance_days")
VALUES (1, 480, 1080, 120, 14);
