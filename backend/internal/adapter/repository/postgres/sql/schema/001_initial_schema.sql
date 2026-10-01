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
  'cancelled',
  'awaiting_payment',
  'expired',
  'no_show'
);
CREATE TYPE "station" AS ENUM ('kitchen', 'counter');
CREATE TYPE "order_type" AS ENUM ('instant', 'pre_order');
CREATE TYPE "ticket_status" AS ENUM ('queued', 'in_progress', 'ready', 'cancelled');

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
  "terms_accepted_at" timestamptz NULL,
  "terms_version" text NULL,
  "erased_at" timestamptz NULL,
  "session_version" integer NOT NULL DEFAULT 0,
  PRIMARY KEY ("id"),
  CONSTRAINT "users_erased_closed_check" CHECK ((erased_at IS NULL) OR (deleted_at IS NOT NULL)),
  CONSTRAINT "users_terms_pair_check" CHECK ((terms_accepted_at IS NULL) = (terms_version IS NULL))
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
  "station" "station" NOT NULL DEFAULT 'kitchen',
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
  "lead_time_minutes" integer NOT NULL DEFAULT 0,
  "is_customizable" boolean NOT NULL DEFAULT false,
  "sold_out_on" date NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "products_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "categories" ("id") ON DELETE RESTRICT,
  CONSTRAINT "products_price_cents_check" CHECK (price_cents >= 0),
  CONSTRAINT "products_lead_time_minutes_check" CHECK ((lead_time_minutes >= 0) AND (lead_time_minutes <= 10080))
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

CREATE TYPE "product_option_group" AS ENUM ('size', 'flavor', 'decoration');

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
  CONSTRAINT "product_options_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "products" ("id") ON DELETE CASCADE,
  CONSTRAINT "product_options_label_check" CHECK ((char_length(label) >= 1) AND (char_length(label) <= 60)),
  CONSTRAINT "product_options_price_delta_cents_check" CHECK ((price_delta_cents >= 0) AND (price_delta_cents <= 100000)),
  CONSTRAINT "product_options_sort_order_check" CHECK (sort_order >= 0)
);
CREATE INDEX "product_options_product_idx" ON "product_options" ("product_id", "option_group", "sort_order") WHERE (deleted_at IS NULL);

CREATE TABLE "combos" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "slug" citext NOT NULL,
  "price_cents" bigint NOT NULL,
  "image_url" text NULL,
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
  "max_uses_per_customer" integer NULL,
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
  CONSTRAINT "discount_codes_max_uses_per_customer_check" CHECK (max_uses_per_customer IS NULL OR max_uses_per_customer > 0),
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
CREATE UNIQUE INDEX "cart_items_cart_plain_product_idx" ON "cart_items" ("cart_id", "product_id") WHERE ("line_type" = 'product' AND "configuration" = '{}'::jsonb);
CREATE UNIQUE INDEX "cart_items_cart_combo_idx" ON "cart_items" ("cart_id", "combo_id") WHERE ("line_type" = 'combo');
CREATE INDEX "cart_items_cart_id_idx" ON "cart_items" ("cart_id");

CREATE TYPE "order_channel" AS ENUM ('online', 'counter', 'phone');

CREATE TABLE "orders" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NULL,
  "status" "order_status" NOT NULL DEFAULT 'pending',
  "subtotal_cents" bigint NOT NULL,
  "discount_cents" bigint NOT NULL DEFAULT 0,
  "total_cents" bigint NOT NULL,
  "discount_code_id" uuid NULL,
  "discount_code_snapshot" text NULL,
  "pickup_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "code" text NOT NULL,
  "order_type" "order_type" NOT NULL,
  "terms_accepted_at" timestamptz NULL,
  "terms_version" text NULL,
  "checkout_key" uuid NULL,
  "payment_due_at" timestamptz NULL,
  "channel" "order_channel" NOT NULL DEFAULT 'online',
  "guest_name" text NULL,
  "guest_phone" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "orders_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE RESTRICT,
  CONSTRAINT "orders_discount_code_id_fkey" FOREIGN KEY ("discount_code_id") REFERENCES "discount_codes" ("id") ON DELETE SET NULL,
  CONSTRAINT "orders_subtotal_cents_check" CHECK (subtotal_cents >= 0),
  CONSTRAINT "orders_discount_cents_check" CHECK (discount_cents >= 0),
  CONSTRAINT "orders_total_cents_check" CHECK (total_cents >= 0),
  CONSTRAINT "orders_total_balance_check" CHECK (total_cents = subtotal_cents - discount_cents),
  CONSTRAINT "orders_code_check" CHECK (code ~ '^CH-[0-9]{6}-[0-9]{3,}$'::text),
  CONSTRAINT "orders_terms_pair_check" CHECK ((terms_accepted_at IS NULL) = (terms_version IS NULL)),
  CONSTRAINT "orders_customer_check" CHECK (((user_id IS NULL) = (guest_name IS NOT NULL)) AND ((guest_name IS NULL) = (guest_phone IS NULL)) AND ((channel <> 'online'::order_channel) OR (user_id IS NOT NULL)))
);
CREATE UNIQUE INDEX "orders_code_idx" ON "orders" ("code");
CREATE UNIQUE INDEX "orders_checkout_key_idx" ON "orders" ("user_id", "checkout_key") WHERE (checkout_key IS NOT NULL);
CREATE UNIQUE INDEX "orders_staff_checkout_key_idx" ON "orders" ("checkout_key") WHERE ((channel <> 'online'::order_channel) AND (checkout_key IS NOT NULL));
CREATE INDEX "orders_payment_due_idx" ON "orders" ("payment_due_at") WHERE (status = 'awaiting_payment'::order_status);
CREATE INDEX "orders_pickup_slot_idx" ON "orders" ("pickup_at") WHERE (status <> 'cancelled'::order_status);
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

CREATE TABLE "order_tickets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "station" "station" NOT NULL,
  "status" "ticket_status" NOT NULL DEFAULT 'queued',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_tickets_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON DELETE CASCADE
);
CREATE UNIQUE INDEX "order_tickets_order_station_idx" ON "order_tickets" ("order_id", "station");
CREATE INDEX "order_tickets_station_queue_idx" ON "order_tickets" ("station", "status") WHERE (status <> 'cancelled'::ticket_status);

CREATE TABLE "order_ticket_items" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "ticket_id" uuid NOT NULL,
  "order_item_id" uuid NOT NULL,
  "product_id" uuid NOT NULL,
  "name" text NOT NULL,
  "quantity" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_ticket_items_order_item_id_fkey" FOREIGN KEY ("order_item_id") REFERENCES "order_items" ("id") ON DELETE CASCADE,
  CONSTRAINT "order_ticket_items_ticket_id_fkey" FOREIGN KEY ("ticket_id") REFERENCES "order_tickets" ("id") ON DELETE CASCADE,
  CONSTRAINT "order_ticket_items_quantity_check" CHECK (quantity > 0)
);
CREATE INDEX "order_ticket_items_ticket_id_idx" ON "order_ticket_items" ("ticket_id");

CREATE TABLE "order_day_counters" (
  "day" date NOT NULL,
  "last_number" integer NOT NULL,
  PRIMARY KEY ("day"),
  CONSTRAINT "order_day_counters_last_number_check" CHECK (last_number > 0)
);

CREATE TABLE "order_status_events" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "from_status" "order_status" NULL,
  "to_status" "order_status" NOT NULL,
  "actor_id" uuid NULL,
  "actor_role" "user_role" NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "reason" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "order_status_events_actor_id_fkey" FOREIGN KEY ("actor_id") REFERENCES "users" ("id") ON DELETE RESTRICT,
  CONSTRAINT "order_status_events_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON DELETE CASCADE,
  CONSTRAINT "order_status_events_move_check" CHECK (from_status IS DISTINCT FROM to_status),
  CONSTRAINT "order_status_events_actor_check" CHECK ((actor_id IS NULL) = (actor_role IS NULL)),
  CONSTRAINT "order_status_events_reason_check" CHECK ((reason IS NULL) OR ((to_status = 'cancelled'::order_status) AND ((char_length(reason) >= 1) AND (char_length(reason) <= 200))))
);
CREATE INDEX "order_status_events_order_created_idx" ON "order_status_events" ("order_id", "created_at");

-- Create "outbox_events" table
CREATE TABLE "outbox_events" (
  "id" uuid NOT NULL,
  "topic" text NOT NULL,
  "audience" jsonb NOT NULL,
  "data" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT clock_timestamp(),
  "published_at" timestamptz NULL,
  "attempts" integer NOT NULL DEFAULT 0,
  "last_error" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "outbox_events_published_idx" to table: "outbox_events"
CREATE INDEX "outbox_events_published_idx" ON "outbox_events" ("published_at") WHERE (published_at IS NOT NULL);
-- Create index "outbox_events_unpublished_idx" to table: "outbox_events"
CREATE INDEX "outbox_events_unpublished_idx" ON "outbox_events" ("created_at") WHERE (published_at IS NULL);
-- Create "store_closed_dates" table
CREATE TABLE "store_closed_dates" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "closed_on" date NOT NULL,
  "reason" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "store_closed_dates_reason_check" CHECK ((char_length(reason) >= 1) AND (char_length(reason) <= 200))
);
-- Create index "store_closed_dates_day_idx" to table: "store_closed_dates"
CREATE UNIQUE INDEX "store_closed_dates_day_idx" ON "store_closed_dates" ("closed_on") WHERE (deleted_at IS NULL);
-- Create "store_settings" table
CREATE TABLE "store_settings" (
  "id" smallint NOT NULL DEFAULT 1,
  "opens_at_minute" smallint NOT NULL,
  "closes_at_minute" smallint NOT NULL,
  "preorder_min_lead_minutes" integer NOT NULL,
  "max_advance_days" smallint NOT NULL,
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "slot_minutes" smallint NOT NULL DEFAULT 30,
  "slot_capacity" smallint NOT NULL DEFAULT 10,
  "instant_prep_minutes" smallint NOT NULL DEFAULT 20,
  "payment_hold_minutes" smallint NOT NULL DEFAULT 15,
  PRIMARY KEY ("id"),
  CONSTRAINT "store_settings_advance_check" CHECK ((max_advance_days >= 1) AND (max_advance_days <= 90)),
  CONSTRAINT "store_settings_capacity_check" CHECK ((slot_capacity >= 1) AND (slot_capacity <= 200)),
  CONSTRAINT "store_settings_instant_prep_check" CHECK ((instant_prep_minutes >= 0) AND (instant_prep_minutes <= 240)),
  CONSTRAINT "store_settings_payment_hold_check" CHECK ((payment_hold_minutes >= 5) AND (payment_hold_minutes <= 120)),
  CONSTRAINT "store_settings_slot_check" CHECK ((slot_minutes = ANY (ARRAY[10, 15, 20, 30, 60])) AND ((closes_at_minute - opens_at_minute) >= slot_minutes)),
  CONSTRAINT "store_settings_hours_check" CHECK ((opens_at_minute >= 0) AND (closes_at_minute < 1440) AND (opens_at_minute < closes_at_minute)),
  CONSTRAINT "store_settings_lead_check" CHECK ((preorder_min_lead_minutes >= 0) AND (preorder_min_lead_minutes <= 10080)),
  CONSTRAINT "store_settings_singleton_check" CHECK (id = 1)
);

CREATE TYPE "user_token_purpose" AS ENUM ('verify_email', 'reset_password');

CREATE TABLE "user_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "purpose" "user_token_purpose" NOT NULL,
  "token_hash" bytea NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "user_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_tokens_token_hash_check" CHECK (octet_length(token_hash) = 32)
);
CREATE UNIQUE INDEX "user_tokens_token_hash_idx" ON "user_tokens" ("token_hash");
CREATE UNIQUE INDEX "user_tokens_user_purpose_idx" ON "user_tokens" ("user_id", "purpose");

CREATE TYPE "payment_provider" AS ENUM ('paypal', 'cash');
CREATE TYPE "payment_status" AS ENUM ('created', 'pending', 'captured', 'denied', 'refunded');

CREATE TABLE "payments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "provider" "payment_provider" NOT NULL,
  "provider_order_id" text NULL,
  "approve_url" text NULL,
  "status" "payment_status" NOT NULL DEFAULT 'created',
  "capture_id" text NULL,
  "amount_cents" bigint NOT NULL,
  "currency" text NOT NULL,
  "captured_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "refund_requested_at" timestamptz NULL,
  "refund_id" text NULL,
  "refunded_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "payments_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON DELETE RESTRICT,
  CONSTRAINT "payments_amount_cents_check" CHECK (amount_cents > 0),
  CONSTRAINT "payments_capture_check" CHECK ((provider <> 'paypal'::payment_provider) OR ((status = 'created'::payment_status) = (capture_id IS NULL))),
  CONSTRAINT "payments_provider_check" CHECK ((provider = 'paypal'::payment_provider) = ((provider_order_id IS NOT NULL) AND (approve_url IS NOT NULL))),
  CONSTRAINT "payments_cash_check" CHECK ((provider <> 'cash'::payment_provider) OR ((capture_id IS NULL) AND (status = ANY (ARRAY['created'::payment_status, 'captured'::payment_status])) AND (refund_requested_at IS NULL))),
  CONSTRAINT "payments_captured_at_check" CHECK ((status = ANY (ARRAY['captured'::payment_status, 'refunded'::payment_status])) = (captured_at IS NOT NULL)),
  CONSTRAINT "payments_currency_check" CHECK (currency ~ '^[A-Z]{3}$'),
  CONSTRAINT "payments_refund_requested_check" CHECK (((refund_requested_at IS NULL) OR (captured_at IS NOT NULL)) AND ((refunded_at IS NULL) OR (refund_requested_at IS NOT NULL))),
  CONSTRAINT "payments_refunded_at_check" CHECK ((status = 'refunded'::payment_status) = (refunded_at IS NOT NULL))
);
CREATE UNIQUE INDEX "payments_capture_idx" ON "payments" ("provider", "capture_id") WHERE (capture_id IS NOT NULL);
CREATE UNIQUE INDEX "payments_order_id_idx" ON "payments" ("order_id");
CREATE UNIQUE INDEX "payments_provider_order_idx" ON "payments" ("provider", "provider_order_id");
CREATE INDEX "payments_refund_due_idx" ON "payments" ("refund_requested_at") WHERE ((status = 'captured'::payment_status) AND (refund_requested_at IS NOT NULL));
