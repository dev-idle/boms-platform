-- Modify "messages" table
ALTER TABLE "messages" ADD COLUMN "via_assistant" boolean NOT NULL DEFAULT false;
