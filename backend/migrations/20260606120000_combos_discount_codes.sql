-- Combos (manager-owned bundles) and discount codes (manager-owned promotions).

CREATE TYPE "discount_type" AS ENUM ('percent', 'fixed_cents');

CREATE TABLE "combos" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "slug" citext NOT NULL,
  "price_cents" bigint NOT NULL,
  "starts_at" timestamptz NOT NULL,
  "ends_at" timestamptz NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "combos_price_cents_check" CHECK (price_cents >= 0),
  CONSTRAINT "combos_window_check" CHECK (ends_at > starts_at)
);
CREATE UNIQUE INDEX "combos_slug_active_idx" ON "combos" ("slug") WHERE (deleted_at IS NULL);
CREATE INDEX "combos_active_window_idx" ON "combos" ("is_active", "starts_at", "ends_at") WHERE (deleted_at IS NULL);

CREATE TABLE "combo_items" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "combo_id" uuid NOT NULL,
  "product_id" uuid NOT NULL,
  "quantity" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "combo_items_combo_id_fkey" FOREIGN KEY ("combo_id") REFERENCES "combos" ("id") ON DELETE CASCADE,
  CONSTRAINT "combo_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE RESTRICT,
  CONSTRAINT "combo_items_quantity_check" CHECK (quantity > 0),
  CONSTRAINT "combo_items_combo_product_unique" UNIQUE ("combo_id", "product_id")
);
CREATE INDEX "combo_items_combo_id_idx" ON "combo_items" ("combo_id");

CREATE TABLE "discount_codes" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "code" citext NOT NULL,
  "discount_type" "discount_type" NOT NULL,
  "value" bigint NOT NULL,
  "min_order_cents" bigint NULL,
  "max_uses" integer NULL,
  "used_count" integer NOT NULL DEFAULT 0,
  "starts_at" timestamptz NOT NULL,
  "ends_at" timestamptz NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "discount_codes_window_check" CHECK (ends_at > starts_at),
  CONSTRAINT "discount_codes_used_count_check" CHECK (used_count >= 0),
  CONSTRAINT "discount_codes_min_order_cents_check" CHECK (min_order_cents IS NULL OR min_order_cents >= 0),
  CONSTRAINT "discount_codes_max_uses_check" CHECK (max_uses IS NULL OR max_uses > 0)
);
CREATE UNIQUE INDEX "discount_codes_code_active_idx" ON "discount_codes" ("code") WHERE (deleted_at IS NULL);
CREATE INDEX "discount_codes_active_window_idx" ON "discount_codes" ("is_active", "starts_at", "ends_at") WHERE (deleted_at IS NULL);
