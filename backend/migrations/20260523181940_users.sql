-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" public.citext NOT NULL,
  "password_hash" text NOT NULL,
  "role" "user_role" NOT NULL DEFAULT 'customer',
  "email_verified_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "users_email_active_idx" to table: "users"
CREATE UNIQUE INDEX "users_email_active_idx" ON "users" ("email") WHERE (deleted_at IS NULL);
-- Create index "users_role_idx" to table: "users"
CREATE INDEX "users_role_idx" ON "users" ("role") WHERE (deleted_at IS NULL);
