-- Customer carts and orders (server-authoritative pricing at read/checkout).

CREATE TYPE "cart_line_type" AS ENUM ('product', 'combo');
CREATE TYPE "order_status" AS ENUM ('pending', 'confirmed', 'cancelled', 'fulfilled');

CREATE TABLE "carts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "discount_code_id" uuid NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "carts_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE,
  CONSTRAINT "carts_discount_code_id_fkey" FOREIGN KEY ("discount_code_id") REFERENCES "discount_codes" ("id") ON DELETE SET NULL
);
CREATE UNIQUE INDEX "carts_user_id_idx" ON "carts" ("user_id");

CREATE TABLE "cart_items" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "cart_id" uuid NOT NULL,
  "line_type" "cart_line_type" NOT NULL,
  "product_id" uuid NULL,
  "combo_id" uuid NULL,
  "quantity" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "cart_items_cart_id_fkey" FOREIGN KEY ("cart_id") REFERENCES "carts" ("id") ON DELETE CASCADE,
  CONSTRAINT "cart_items_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE RESTRICT,
  CONSTRAINT "cart_items_combo_id_fkey" FOREIGN KEY ("combo_id") REFERENCES "combos" ("id") ON DELETE RESTRICT,
  CONSTRAINT "cart_items_quantity_check" CHECK (quantity > 0),
  CONSTRAINT "cart_items_line_target_check" CHECK (
    ("line_type" = 'product' AND "product_id" IS NOT NULL AND "combo_id" IS NULL)
    OR ("line_type" = 'combo' AND "combo_id" IS NOT NULL AND "product_id" IS NULL)
  )
);
CREATE UNIQUE INDEX "cart_items_cart_product_idx" ON "cart_items" ("cart_id", "product_id") WHERE ("line_type" = 'product');
CREATE UNIQUE INDEX "cart_items_cart_combo_idx" ON "cart_items" ("cart_id", "combo_id") WHERE ("line_type" = 'combo');
CREATE INDEX "cart_items_cart_id_idx" ON "cart_items" ("cart_id");

CREATE TABLE "orders" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "status" "order_status" NOT NULL DEFAULT 'pending',
  "subtotal_cents" bigint NOT NULL,
  "discount_cents" bigint NOT NULL DEFAULT 0,
  "total_cents" bigint NOT NULL,
  "discount_code_id" uuid NULL,
  "discount_code_snapshot" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "orders_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE RESTRICT,
  CONSTRAINT "orders_discount_code_id_fkey" FOREIGN KEY ("discount_code_id") REFERENCES "discount_codes" ("id") ON DELETE SET NULL,
  CONSTRAINT "orders_subtotal_cents_check" CHECK (subtotal_cents >= 0),
  CONSTRAINT "orders_discount_cents_check" CHECK (discount_cents >= 0),
  CONSTRAINT "orders_total_cents_check" CHECK (total_cents >= 0),
  CONSTRAINT "orders_total_balance_check" CHECK (total_cents = subtotal_cents - discount_cents)
);
CREATE INDEX "orders_user_id_created_at_idx" ON "orders" ("user_id", "created_at" DESC);

CREATE TABLE "order_items" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "line_type" "cart_line_type" NOT NULL,
  "product_id" uuid NULL,
  "combo_id" uuid NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "quantity" integer NOT NULL,
  "unit_price_cents" bigint NOT NULL,
  "line_total_cents" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_items_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON DELETE CASCADE,
  CONSTRAINT "order_items_quantity_check" CHECK (quantity > 0),
  CONSTRAINT "order_items_unit_price_cents_check" CHECK (unit_price_cents >= 0),
  CONSTRAINT "order_items_line_total_cents_check" CHECK (line_total_cents >= 0),
  CONSTRAINT "order_items_line_target_check" CHECK (
    ("line_type" = 'product' AND "product_id" IS NOT NULL AND "combo_id" IS NULL)
    OR ("line_type" = 'combo' AND "combo_id" IS NOT NULL AND "product_id" IS NULL)
  )
);
CREATE INDEX "order_items_order_id_idx" ON "order_items" ("order_id");
