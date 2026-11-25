-- Modify "conversations" table
ALTER TABLE "conversations" ADD COLUMN "reply_emailed_at" timestamptz NULL;
-- Create "policy_versions" table
CREATE TABLE "policy_versions" (
  "version" text NOT NULL,
  "published_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("version"),
  CONSTRAINT "policy_versions_version_check" CHECK ((char_length(version) >= 1) AND (char_length(version) <= 32))
);
-- Create "reference_photo_erasures" table
CREATE TABLE "reference_photo_erasures" (
  "user_id" uuid NOT NULL,
  "requested_at" timestamptz NOT NULL DEFAULT now(),
  "erased_at" timestamptz NULL,
  PRIMARY KEY ("user_id"),
  CONSTRAINT "reference_photo_erasures_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- Create index "reference_photo_erasures_due_idx" to table: "reference_photo_erasures"
CREATE INDEX "reference_photo_erasures_due_idx" ON "reference_photo_erasures" ("requested_at") WHERE (erased_at IS NULL);
-- Customers erased before now: Cloudinary still holds their reference photos,
-- and their order lines still name the photo and the cake's message.
INSERT INTO "reference_photo_erasures" ("user_id")
SELECT "id" FROM "users" WHERE "role" = 'customer' AND "erased_at" IS NOT NULL;
UPDATE "order_items"
SET "configuration" = "configuration" - 'reference_image_url' - 'message'
WHERE ("configuration"->>'reference_image_url' IS NOT NULL OR "configuration"->>'message' IS NOT NULL)
  AND "order_id" IN (SELECT o."id" FROM "orders" o JOIN "users" u ON u."id" = o."user_id" WHERE u."erased_at" IS NOT NULL);
-- Replies left unread before now are not emailed about: the email is for
-- replies that wait from now on.
UPDATE "conversations" SET "reply_emailed_at" = now() WHERE "unread_by_customer" > 0;
