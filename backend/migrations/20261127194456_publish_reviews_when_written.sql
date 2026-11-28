-- A review shows from the moment it is written and a manager may hide it, so
-- nothing waits for a manager any more: "pending" leaves the enum. Postgres
-- cannot drop an enum value, so the type is replaced; that rewrites "reviews"
-- under an exclusive lock, which the size of the table allows.
-- Publish the reviews still waiting
UPDATE "reviews" SET "status" = 'published' WHERE "status" = 'pending';
-- Drop what names the old type
ALTER TABLE "reviews" DROP CONSTRAINT "reviews_moderation_check";
DROP INDEX "reviews_product_published_idx";
DROP INDEX "reviews_status_created_idx";
ALTER TABLE "reviews" ALTER COLUMN "status" DROP DEFAULT;
-- Replace enum type "review_status"
ALTER TYPE "review_status" RENAME TO "review_status_old";
CREATE TYPE "review_status" AS ENUM ('published', 'hidden');
ALTER TABLE "reviews" ALTER COLUMN "status" TYPE "review_status" USING ("status"::text::"review_status");
DROP TYPE "review_status_old";
-- Modify "reviews" table
ALTER TABLE "reviews" ALTER COLUMN "status" SET DEFAULT 'published';
ALTER TABLE "reviews" ADD CONSTRAINT "reviews_moderation_check" CHECK (((moderated_by IS NULL) = (moderated_at IS NULL)) AND ((status = 'published'::review_status) OR (moderated_at IS NOT NULL)));
-- Create index "reviews_product_published_idx" to table: "reviews"
CREATE INDEX "reviews_product_published_idx" ON "reviews" ("product_id", "created_at") WHERE ((status = 'published'::review_status) AND (deleted_at IS NULL));
-- Create index "reviews_status_created_idx" to table: "reviews"
CREATE INDEX "reviews_status_created_idx" ON "reviews" ("status", "created_at") WHERE (deleted_at IS NULL);
