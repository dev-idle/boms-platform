-- Create enum type "saved_list"
CREATE TYPE "saved_list" AS ENUM ('favorite', 'wishlist');
-- Create "saved_products" table
CREATE TABLE "saved_products" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "product_id" uuid NOT NULL,
  "list" "saved_list" NOT NULL,
  "saved_at" timestamptz NOT NULL DEFAULT now(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "saved_products_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "saved_products_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- Create index "saved_products_user_list_saved_idx" to table: "saved_products"
CREATE INDEX "saved_products_user_list_saved_idx" ON "saved_products" ("user_id", "list", "saved_at") WHERE (deleted_at IS NULL);
-- Create index "saved_products_user_product_list_idx" to table: "saved_products"
CREATE UNIQUE INDEX "saved_products_user_product_list_idx" ON "saved_products" ("user_id", "product_id", "list");
