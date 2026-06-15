-- Product gallery: up to 5 images per product (replaces products.image_url).

CREATE TABLE "product_images" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "product_id" uuid NOT NULL,
  "sort_order" smallint NOT NULL,
  "image_url" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "product_images_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE,
  CONSTRAINT "product_images_sort_order_range" CHECK (sort_order >= 0 AND sort_order < 5),
  CONSTRAINT "product_images_product_sort_unique" UNIQUE ("product_id", "sort_order")
);

CREATE INDEX "product_images_product_id_sort_idx" ON "product_images" ("product_id", "sort_order");

INSERT INTO "product_images" ("product_id", "sort_order", "image_url")
SELECT "id", 0, trim("image_url")
FROM "products"
WHERE "image_url" IS NOT NULL
  AND trim("image_url") <> '';

ALTER TABLE "products" DROP COLUMN "image_url";
