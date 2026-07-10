-- BOMS schema v2 baseline (academic clean slate).
-- Enum label order is fixed at creation; matches db/schema.hcl.
CREATE EXTENSION IF NOT EXISTS "citext" WITH SCHEMA "public" VERSION "1.6";
CREATE EXTENSION IF NOT EXISTS "pgcrypto" WITH SCHEMA "public" VERSION "1.3";

CREATE TYPE "user_role" AS ENUM ('admin', 'customer', 'staff', 'baker', 'manager');
CREATE TYPE "discount_type" AS ENUM ('percent', 'fixed_cents');
CREATE TYPE "line_type" AS ENUM ('product', 'combo');
CREATE TYPE "order_status" AS ENUM (
  'pending',
  'confirmed',
  'in_production',
  'ready',
  'fulfilled',
  'cancelled'
);

CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "email" public.citext NOT NULL,
  "password_hash" text NOT NULL,
  "role" "user_role" NOT NULL DEFAULT 'customer',
  "email_verified_at" timestamptz NULL,
  "must_change_password" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
CREATE UNIQUE INDEX "users_email_active_idx" ON "users" ("email") WHERE (deleted_at IS NULL);
CREATE INDEX "users_role_idx" ON "users" ("role") WHERE (deleted_at IS NULL);

CREATE TABLE "customer_profiles" (
  "user_id" uuid NOT NULL,
  "display_name" text NULL,
  "phone" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("user_id"),
  CONSTRAINT "customer_profiles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE "staff_profiles" (
  "user_id" uuid NOT NULL,
  "full_name" text NOT NULL DEFAULT '',
  "phone" text NULL,
  "employee_code" citext NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("user_id"),
  CONSTRAINT "staff_profiles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX "staff_profiles_employee_code_idx" ON "staff_profiles" ("employee_code");

CREATE TABLE "admin_profiles" (
  "user_id" uuid NOT NULL,
  "full_name" text NOT NULL DEFAULT '',
  "phone" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("user_id"),
  CONSTRAINT "admin_profiles_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE
);

CREATE TABLE "audit_logs" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "actor_id" uuid NOT NULL,
  "actor_role" "user_role" NOT NULL,
  "action" text NOT NULL,
  "target_id" uuid NULL,
  "target_type" text NOT NULL,
  "before_jsonb" jsonb NOT NULL DEFAULT '{}',
  "after_jsonb" jsonb NOT NULL DEFAULT '{}',
  "ip" inet NULL,
  "user_agent" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "audit_logs_actor_id_fkey" FOREIGN KEY ("actor_id") REFERENCES "users" ("id") ON DELETE RESTRICT
);
CREATE INDEX "audit_logs_actor_created_idx" ON "audit_logs" ("actor_id", "created_at" DESC);
CREATE INDEX "audit_logs_target_idx" ON "audit_logs" ("target_type", "target_id");

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
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "products_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE RESTRICT,
  CONSTRAINT "products_price_cents_check" CHECK (price_cents >= 0)
);
CREATE UNIQUE INDEX "products_slug_active_idx" ON "products" ("slug") WHERE (deleted_at IS NULL);
CREATE INDEX "products_category_active_idx" ON "products" ("category_id") WHERE (deleted_at IS NULL);

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
  "updated_at" timestamptz NOT NULL DEFAULT now(),
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
  "max_discount_cents" bigint NULL,
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
  CONSTRAINT "discount_codes_max_uses_check" CHECK (max_uses IS NULL OR max_uses > 0),
  CONSTRAINT "discount_codes_value_percent_check" CHECK (discount_type <> 'percent' OR (value >= 1 AND value <= 100)),
  CONSTRAINT "discount_codes_value_fixed_cents_check" CHECK (discount_type <> 'fixed_cents' OR value >= 1),
  CONSTRAINT "discount_codes_used_within_max_check" CHECK (max_uses IS NULL OR used_count <= max_uses),
  CONSTRAINT "discount_codes_max_discount_cents_check" CHECK (max_discount_cents IS NULL OR max_discount_cents >= 1),
  CONSTRAINT "discount_codes_max_discount_percent_only_check" CHECK (discount_type = 'percent' OR max_discount_cents IS NULL)
);
CREATE UNIQUE INDEX "discount_codes_code_active_idx" ON "discount_codes" ("code") WHERE (deleted_at IS NULL);
CREATE INDEX "discount_codes_active_window_idx" ON "discount_codes" ("is_active", "starts_at", "ends_at") WHERE (deleted_at IS NULL);

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
  "line_type" "line_type" NOT NULL,
  "product_id" uuid NULL,
  "combo_id" uuid NULL,
  "quantity" integer NOT NULL,
  "configuration" jsonb NOT NULL DEFAULT '{}',
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
  "pickup_at" timestamptz NULL,
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
CREATE INDEX "orders_production_pickup_idx" ON "orders" ("status", "pickup_at") WHERE (
  status IN ('confirmed'::order_status, 'in_production'::order_status, 'ready'::order_status)
);

CREATE TABLE "order_items" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "line_type" "line_type" NOT NULL,
  "product_id" uuid NULL,
  "combo_id" uuid NULL,
  "configuration" jsonb NOT NULL DEFAULT '{}',
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
