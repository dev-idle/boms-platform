-- Create enum type "review_status"
CREATE TYPE "review_status" AS ENUM ('pending', 'published', 'hidden');
-- Create "reviews" table
CREATE TABLE "reviews" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "product_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "rating" smallint NOT NULL,
  "comment" text NULL,
  "status" "review_status" NOT NULL DEFAULT 'pending',
  "moderated_by" uuid NULL,
  "moderated_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "reviews_moderated_by_fkey" FOREIGN KEY ("moderated_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "reviews_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "reviews_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "reviews_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "reviews_comment_check" CHECK ((comment IS NULL) OR ((char_length(comment) >= 1) AND (char_length(comment) <= 1000))),
  CONSTRAINT "reviews_moderation_check" CHECK (((moderated_by IS NULL) = (moderated_at IS NULL)) AND ((status = 'pending'::review_status) = (moderated_at IS NULL))),
  CONSTRAINT "reviews_rating_check" CHECK ((rating >= 1) AND (rating <= 5))
);
-- Create index "reviews_order_product_idx" to table: "reviews"
CREATE UNIQUE INDEX "reviews_order_product_idx" ON "reviews" ("order_id", "product_id");
-- Create index "reviews_product_published_idx" to table: "reviews"
CREATE INDEX "reviews_product_published_idx" ON "reviews" ("product_id", "created_at") WHERE ((status = 'published'::review_status) AND (deleted_at IS NULL));
-- Create index "reviews_status_created_idx" to table: "reviews"
CREATE INDEX "reviews_status_created_idx" ON "reviews" ("status", "created_at") WHERE (deleted_at IS NULL);
-- Create index "reviews_user_id_idx" to table: "reviews"
CREATE INDEX "reviews_user_id_idx" ON "reviews" ("user_id");
