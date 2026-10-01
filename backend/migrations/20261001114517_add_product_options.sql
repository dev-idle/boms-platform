-- Create enum type "product_option_group"
CREATE TYPE "product_option_group" AS ENUM ('size', 'flavor', 'decoration');
-- Modify "products" table
ALTER TABLE "products" ADD COLUMN "is_customizable" boolean NOT NULL DEFAULT false;
-- Create "product_options" table
CREATE TABLE "product_options" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "product_id" uuid NOT NULL,
  "option_group" "product_option_group" NOT NULL,
  "label" text NOT NULL,
  "price_delta_cents" bigint NOT NULL DEFAULT 0,
  "sort_order" smallint NOT NULL DEFAULT 0,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "product_options_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "product_options_label_check" CHECK ((char_length(label) >= 1) AND (char_length(label) <= 60)),
  CONSTRAINT "product_options_price_delta_cents_check" CHECK ((price_delta_cents >= 0) AND (price_delta_cents <= 100000)),
  CONSTRAINT "product_options_sort_order_check" CHECK (sort_order >= 0)
);
-- Create index "product_options_product_idx" to table: "product_options"
CREATE INDEX "product_options_product_idx" ON "product_options" ("product_id", "option_group", "sort_order") WHERE (deleted_at IS NULL);
