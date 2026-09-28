-- Create "outbox_events" table
CREATE TABLE "outbox_events" (
  "id" uuid NOT NULL,
  "topic" text NOT NULL,
  "audience" jsonb NOT NULL,
  "data" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT clock_timestamp(),
  "published_at" timestamptz NULL,
  "attempts" integer NOT NULL DEFAULT 0,
  "last_error" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "outbox_events_published_idx" to table: "outbox_events"
CREATE INDEX "outbox_events_published_idx" ON "outbox_events" ("published_at") WHERE (published_at IS NOT NULL);
-- Create index "outbox_events_unpublished_idx" to table: "outbox_events"
CREATE INDEX "outbox_events_unpublished_idx" ON "outbox_events" ("created_at") WHERE (published_at IS NULL);
