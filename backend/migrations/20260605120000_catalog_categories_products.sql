-- Catalog tables: categories and products (manager-owned, customer read-only via /catalog/*).

CREATE TABLE "categories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "slug" citext NOT NULL,
  "sort_order" integer NOT NULL DEFAULT 0,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "categories_slug_active_idx" ON "categories" ("slug") WHERE (deleted_at IS NULL);
CREATE INDEX "categories_active_sort_idx" ON "categories" ("is_active", "sort_order") WHERE (deleted_at IS NULL);

CREATE TABLE "products" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "category_id" uuid NOT NULL,
  "name" text NOT NULL,
  "slug" citext NOT NULL,
  "description" text NULL,
  "price_cents" bigint NOT NULL,
  "is_available" boolean NOT NULL DEFAULT true,
  "image_url" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "products_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE RESTRICT,
  CONSTRAINT "products_price_cents_check" CHECK (price_cents >= 0)
);
CREATE UNIQUE INDEX "products_slug_active_idx" ON "products" ("slug") WHERE (deleted_at IS NULL);
CREATE INDEX "products_category_active_idx" ON "products" ("category_id") WHERE (deleted_at IS NULL);
