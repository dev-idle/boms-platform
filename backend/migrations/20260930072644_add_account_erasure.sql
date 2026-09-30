-- Modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_erased_closed_check" CHECK ((erased_at IS NULL) OR (deleted_at IS NOT NULL)), ADD COLUMN "erased_at" timestamptz NULL;
