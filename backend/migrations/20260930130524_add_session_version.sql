-- Modify "users" table
ALTER TABLE "users" ADD COLUMN "session_version" integer NOT NULL DEFAULT 0;
