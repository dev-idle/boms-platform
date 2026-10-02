// Declarative schema (source of truth). Keep sql/schema/ aligned for sqlc after each migration.
schema "public" {}

extension "citext" {
  schema  = schema.public
  version = "1.6"
}

extension "pgcrypto" {
  schema  = schema.public
  version = "1.3"
}

enum "user_role" {
  schema = schema.public
  values = ["admin", "customer", "staff", "baker", "manager"]
}

table "users" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "email" {
    type = sql("citext")
    null = false
  }
  column "password_hash" {
    type = text
    null = false
  }
  column "role" {
    type    = enum.user_role
    null    = false
    default = sql("'customer'::user_role")
  }
  column "email_verified_at" {
    type = timestamptz
    null = true
  }
  column "must_change_password" {
    type    = boolean
    null    = false
    default = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  column "terms_accepted_at" {
    type = timestamptz
    null = true
  }
  column "terms_version" {
    type = text
    null = true
  }
  column "erased_at" {
    type = timestamptz
    null = true
  }
  column "session_version" {
    type    = integer
    null    = false
    default = 0
  }
  primary_key {
    columns = [column.id]
  }
  check "users_erased_closed_check" {
    expr = "erased_at IS NULL OR deleted_at IS NOT NULL"
  }
  check "users_terms_pair_check" {
    expr = "(terms_accepted_at IS NULL) = (terms_version IS NULL)"
  }
  index "users_email_active_idx" {
    unique  = true
    columns = [column.email]
    where   = "deleted_at IS NULL"
  }
  index "users_role_idx" {
    columns = [column.role]
    where   = "deleted_at IS NULL"
  }
}

table "customer_profiles" {
  schema = schema.public
  column "user_id" {
    type = uuid
    null = false
  }
  column "display_name" {
    type = text
    null = true
  }
  column "phone" {
    type = text
    null = true
  }
  // When the customer agreed to be emailed promotions; empty while they have
  // not, or since they withdrew.
  column "marketing_consent_at" {
    type = timestamptz
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.user_id]
  }
  foreign_key "customer_profiles_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = CASCADE
  }
}

table "staff_profiles" {
  schema = schema.public
  column "user_id" {
    type = uuid
    null = false
  }
  column "full_name" {
    type    = text
    null    = false
    default = ""
  }
  column "phone" {
    type = text
    null = true
  }
  column "employee_code" {
    type = sql("citext")
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.user_id]
  }
  foreign_key "staff_profiles_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = CASCADE
  }
  index "staff_profiles_employee_code_idx" {
    unique  = true
    columns = [column.employee_code]
  }
}

table "admin_profiles" {
  schema = schema.public
  column "user_id" {
    type = uuid
    null = false
  }
  column "full_name" {
    type    = text
    null    = false
    default = ""
  }
  column "phone" {
    type = text
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.user_id]
  }
  foreign_key "admin_profiles_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = CASCADE
  }
}

table "audit_logs" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "actor_id" {
    type = uuid
    null = false
  }
  column "actor_role" {
    type = enum.user_role
    null = false
  }
  column "action" {
    type = text
    null = false
  }
  column "target_id" {
    type = uuid
    null = true
  }
  column "target_type" {
    type = text
    null = false
  }
  column "before_jsonb" {
    type    = jsonb
    null    = false
    default = sql("'{}'::jsonb")
  }
  column "after_jsonb" {
    type    = jsonb
    null    = false
    default = sql("'{}'::jsonb")
  }
  column "ip" {
    type = sql("inet")
    null = true
  }
  column "user_agent" {
    type = text
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "audit_logs_actor_id_fkey" {
    columns     = [column.actor_id]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  index "audit_logs_actor_created_idx" {
    on {
      column = column.actor_id
    }
    on {
      column = column.created_at
      desc   = true
    }
  }
  index "audit_logs_target_idx" {
    columns = [column.target_type, column.target_id]
  }
}

// Where a category's products come from — the kitchen or the counter.
enum "station" {
  schema = schema.public
  values = ["kitchen", "counter"]
}

table "categories" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "name" {
    type = text
    null = false
  }
  column "slug" {
    type = sql("citext")
    null = false
  }
  column "sort_order" {
    type    = int
    null    = false
    default = 0
  }
  column "is_active" {
    type    = boolean
    null    = false
    default = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  // Where the category's products are made: the kitchen bakes to order, the
  // counter sells what is ready.
  column "station" {
    type    = enum.station
    null    = false
    default = sql("'kitchen'::station")
  }
  primary_key {
    columns = [column.id]
  }
  index "categories_slug_active_idx" {
    unique  = true
    columns = [column.slug]
    where   = "deleted_at IS NULL"
  }
  index "categories_active_sort_idx" {
    columns = [column.is_active, column.sort_order]
    where   = "deleted_at IS NULL"
  }
}

table "products" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "category_id" {
    type = uuid
    null = false
  }
  column "name" {
    type = text
    null = false
  }
  column "slug" {
    type = sql("citext")
    null = false
  }
  column "description" {
    type = text
    null = true
  }
  column "price_cents" {
    type = bigint
    null = false
  }
  column "is_active" {
    type    = boolean
    null    = false
    default = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  // The notice this product needs before pickup, on top of the bakery's.
  column "lead_time_minutes" {
    type    = int
    null    = false
    default = 0
  }
  // A customer configures it from its options before it goes in the cart.
  column "is_customizable" {
    type    = boolean
    null    = false
    default = false
  }
  // The bakery day the counter ran out of it: no order collected that day
  // may hold it. A later day is unaffected, so it lapses by itself.
  column "sold_out_on" {
    type = date
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "products_category_id_fkey" {
    columns     = [column.category_id]
    ref_columns = [table.categories.column.id]
    on_delete   = RESTRICT
  }
  index "products_slug_active_idx" {
    unique  = true
    columns = [column.slug]
    where   = "deleted_at IS NULL"
  }
  index "products_category_active_idx" {
    columns = [column.category_id]
    where   = "deleted_at IS NULL"
  }
  check "products_price_cents_check" {
    expr = "price_cents >= 0"
  }
  check "products_lead_time_minutes_check" {
    expr = "lead_time_minutes >= 0 AND lead_time_minutes <= 10080"
  }
}

table "product_images" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "product_id" {
    type = uuid
    null = false
  }
  column "sort_order" {
    type = smallint
    null = false
  }
  column "image_url" {
    type = text
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "product_images_product_id_fkey" {
    columns     = [column.product_id]
    ref_columns = [table.products.column.id]
    on_delete   = CASCADE
  }
  index "product_images_product_id_sort_idx" {
    columns = [column.product_id, column.sort_order]
  }
  unique "product_images_product_sort_unique" {
    columns = [column.product_id, column.sort_order]
  }
  check "product_images_sort_order_range" {
    expr = "sort_order >= 0 AND sort_order < 5"
  }
}

enum "product_option_group" {
  schema = schema.public
  values = ["size", "flavor", "decoration"]
}

// A choice a customer makes on a customizable product, priced on top of it.
// An option is retired with deleted_at, never removed: cart lines name it.
table "product_options" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "product_id" {
    type = uuid
    null = false
  }
  column "option_group" {
    type = enum.product_option_group
    null = false
  }
  column "label" {
    type = text
    null = false
  }
  column "price_delta_cents" {
    type    = bigint
    null    = false
    default = 0
  }
  column "sort_order" {
    type    = smallint
    null    = false
    default = 0
  }
  column "is_active" {
    type    = boolean
    null    = false
    default = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "product_options_product_id_fkey" {
    columns     = [column.product_id]
    ref_columns = [table.products.column.id]
    on_delete   = CASCADE
  }
  index "product_options_product_idx" {
    columns = [column.product_id, column.option_group, column.sort_order]
    where   = "deleted_at IS NULL"
  }
  check "product_options_label_check" {
    expr = "char_length(label) >= 1 AND char_length(label) <= 60"
  }
  check "product_options_price_delta_cents_check" {
    expr = "price_delta_cents >= 0 AND price_delta_cents <= 100000"
  }
  check "product_options_sort_order_check" {
    expr = "sort_order >= 0"
  }
}

enum "discount_type" {
  schema = schema.public
  values = ["percent", "fixed_cents"]
}

table "combos" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "name" {
    type = text
    null = false
  }
  column "slug" {
    type = sql("citext")
    null = false
  }
  column "price_cents" {
    type = bigint
    null = false
  }
  # One promotional shot of the assembled bundle. Products keep their own
  # image table; a combo is a dated campaign and needs exactly one photo.
  column "image_url" {
    type = text
    null = true
  }
  column "starts_at" {
    type = timestamptz
    null = false
  }
  column "ends_at" {
    type = timestamptz
    null = false
  }
  column "is_active" {
    type    = boolean
    null    = false
    default = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  index "combos_slug_active_idx" {
    unique  = true
    columns = [column.slug]
    where   = "deleted_at IS NULL"
  }
  index "combos_active_window_idx" {
    columns = [column.is_active, column.starts_at, column.ends_at]
    where   = "deleted_at IS NULL"
  }
  check "combos_price_cents_check" {
    expr = "price_cents >= 0"
  }
  check "combos_window_check" {
    expr = "ends_at > starts_at"
  }
}

table "combo_items" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "combo_id" {
    type = uuid
    null = false
  }
  column "product_id" {
    type = uuid
    null = false
  }
  column "quantity" {
    type = int
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "combo_items_combo_id_fkey" {
    columns     = [column.combo_id]
    ref_columns = [table.combos.column.id]
    on_delete   = CASCADE
  }
  foreign_key "combo_items_product_id_fkey" {
    columns     = [column.product_id]
    ref_columns = [table.products.column.id]
    on_delete   = RESTRICT
  }
  index "combo_items_combo_id_idx" {
    columns = [column.combo_id]
  }
  check "combo_items_quantity_check" {
    expr = "quantity > 0"
  }
  unique "combo_items_combo_product_unique" {
    columns = [column.combo_id, column.product_id]
  }
}

table "discount_codes" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "code" {
    type = sql("citext")
    null = false
  }
  column "discount_type" {
    type = enum.discount_type
    null = false
  }
  column "value" {
    type = bigint
    null = false
  }
  column "min_order_cents" {
    type = bigint
    null = true
  }
  column "max_uses" {
    type = int
    null = true
  }
  // How many of one customer's orders not cancelled may use the code.
  column "max_uses_per_customer" {
    type = int
    null = true
  }
  column "max_discount_cents" {
    type = bigint
    null = true
  }
  column "used_count" {
    type    = int
    null    = false
    default = 0
  }
  column "starts_at" {
    type = timestamptz
    null = false
  }
  column "ends_at" {
    type = timestamptz
    null = false
  }
  column "is_active" {
    type    = boolean
    null    = false
    default = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  index "discount_codes_code_active_idx" {
    unique  = true
    columns = [column.code]
    where   = "deleted_at IS NULL"
  }
  index "discount_codes_active_window_idx" {
    columns = [column.is_active, column.starts_at, column.ends_at]
    where   = "deleted_at IS NULL"
  }
  check "discount_codes_window_check" {
    expr = "ends_at > starts_at"
  }
  check "discount_codes_used_count_check" {
    expr = "used_count >= 0"
  }
  check "discount_codes_min_order_cents_check" {
    expr = "min_order_cents IS NULL OR min_order_cents >= 0"
  }
  check "discount_codes_max_uses_check" {
    expr = "max_uses IS NULL OR max_uses > 0"
  }
  check "discount_codes_max_uses_per_customer_check" {
    expr = "max_uses_per_customer IS NULL OR max_uses_per_customer > 0"
  }
  check "discount_codes_value_percent_check" {
    expr = "discount_type <> 'percent' OR (value >= 1 AND value <= 100)"
  }
  check "discount_codes_value_fixed_cents_check" {
    expr = "discount_type <> 'fixed_cents' OR value >= 1"
  }
  check "discount_codes_used_within_max_check" {
    expr = "max_uses IS NULL OR used_count <= max_uses"
  }
  check "discount_codes_max_discount_cents_check" {
    expr = "max_discount_cents IS NULL OR max_discount_cents >= 1"
  }
  check "discount_codes_max_discount_percent_only_check" {
    expr = "discount_type = 'percent' OR max_discount_cents IS NULL"
  }
}

enum "line_type" {
  schema = schema.public
  values = ["product", "combo"]
}

enum "order_status" {
  schema = schema.public
  values = ["pending", "confirmed", "in_production", "ready", "fulfilled", "cancelled", "awaiting_payment", "expired", "no_show"]
}

// Where an order was taken: placed online by its customer, or by staff at the
// counter or on the phone.
enum "order_channel" {
  schema = schema.public
  values = ["online", "counter", "phone"]
}

// How an order is prepared: instant (ready-made, collected the same day) or a pre-order.
enum "order_type" {
  schema = schema.public
  values = ["instant", "pre_order"]
}

// Where one station's part of an order stands.
enum "ticket_status" {
  schema = schema.public
  values = ["queued", "in_progress", "ready", "cancelled"]
}

table "carts" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "discount_code_id" {
    type = uuid
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "carts_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = CASCADE
  }
  foreign_key "carts_discount_code_id_fkey" {
    columns     = [column.discount_code_id]
    ref_columns = [table.discount_codes.column.id]
    on_delete   = SET_NULL
  }
  index "carts_user_id_idx" {
    unique  = true
    columns = [column.user_id]
  }
}

table "cart_items" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "cart_id" {
    type = uuid
    null = false
  }
  column "line_type" {
    type = enum.line_type
    null = false
  }
  column "product_id" {
    type = uuid
    null = true
  }
  column "combo_id" {
    type = uuid
    null = true
  }
  column "quantity" {
    type = int
    null = false
  }
  column "configuration" {
    type    = jsonb
    null    = false
    default = sql("'{}'::jsonb")
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "cart_items_cart_id_fkey" {
    columns     = [column.cart_id]
    ref_columns = [table.carts.column.id]
    on_delete   = CASCADE
  }
  foreign_key "cart_items_product_id_fkey" {
    columns     = [column.product_id]
    ref_columns = [table.products.column.id]
    on_delete   = RESTRICT
  }
  foreign_key "cart_items_combo_id_fkey" {
    columns     = [column.combo_id]
    ref_columns = [table.combos.column.id]
    on_delete   = RESTRICT
  }
  index "cart_items_cart_id_idx" {
    columns = [column.cart_id]
  }
  // One line per plain product; a configured product gets a line per configuration.
  index "cart_items_cart_plain_product_idx" {
    unique  = true
    columns = [column.cart_id, column.product_id]
    where   = "line_type = 'product' AND configuration = '{}'::jsonb"
  }
  index "cart_items_cart_combo_idx" {
    unique  = true
    columns = [column.cart_id, column.combo_id]
    where   = "line_type = 'combo'"
  }
  check "cart_items_quantity_check" {
    expr = "quantity > 0"
  }
  check "cart_items_line_target_check" {
    expr = "(line_type = 'product' AND product_id IS NOT NULL AND combo_id IS NULL) OR (line_type = 'combo' AND combo_id IS NOT NULL AND product_id IS NULL)"
  }
}

table "orders" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  // Empty for a guest without an account, who staff took the order for.
  column "user_id" {
    type = uuid
    null = true
  }
  column "status" {
    type    = enum.order_status
    null    = false
    default = sql("'pending'::order_status")
  }
  column "subtotal_cents" {
    type = bigint
    null = false
  }
  column "discount_cents" {
    type    = bigint
    null    = false
    default = 0
  }
  column "total_cents" {
    type = bigint
    null = false
  }
  column "discount_code_id" {
    type = uuid
    null = true
  }
  column "discount_code_snapshot" {
    type = text
    null = true
  }
  column "pickup_at" {
    type = timestamptz
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  // CH-YYMMDD-NNN: the bakery day the order was placed and its number that day.
  column "code" {
    type = text
    null = false
  }
  column "order_type" {
    type = enum.order_type
    null = false
  }
  column "terms_accepted_at" {
    type = timestamptz
    null = true
  }
  column "terms_version" {
    type = text
    null = true
  }
  // The Idempotency-Key of the checkout that placed the order: the same key
  // from the same customer returns this order instead of placing another.
  column "checkout_key" {
    type = uuid
    null = true
  }
  // When an order awaiting payment expires unpaid, on the database clock.
  column "payment_due_at" {
    type = timestamptz
    null = true
  }
  column "channel" {
    type    = enum.order_channel
    null    = false
    default = "online"
  }
  // Who collects a guest's order and how to reach them.
  column "guest_name" {
    type = text
    null = true
  }
  column "guest_phone" {
    type = text
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "orders_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  foreign_key "orders_discount_code_id_fkey" {
    columns     = [column.discount_code_id]
    ref_columns = [table.discount_codes.column.id]
    on_delete   = SET_NULL
  }
  index "orders_user_id_created_at_idx" {
    on {
      column = column.user_id
    }
    on {
      column = column.created_at
      desc   = true
    }
  }
  // Orders still to make or hand over, by status and pickup.
  index "orders_open_pickup_idx" {
    columns = [column.status, column.pickup_at]
    where   = "status IN ('pending'::order_status, 'confirmed'::order_status, 'in_production'::order_status, 'ready'::order_status)"
  }
  index "orders_code_idx" {
    unique  = true
    columns = [column.code]
  }
  // Orders waiting for payment, found by when they expire.
  index "orders_payment_due_idx" {
    columns = [column.payment_due_at]
    where   = "status = 'awaiting_payment'::order_status"
  }
  index "orders_checkout_key_idx" {
    unique  = true
    columns = [column.user_id, column.checkout_key]
    where   = "checkout_key IS NOT NULL"
  }
  // The Idempotency-Key of an order staff took: the same key returns it.
  index "orders_staff_checkout_key_idx" {
    unique  = true
    columns = [column.checkout_key]
    where   = "channel <> 'online'::order_channel AND checkout_key IS NOT NULL"
  }
  // Orders holding a pickup slot, counted when a checkout takes one.
  index "orders_pickup_slot_idx" {
    columns = [column.pickup_at]
    where   = "status <> 'cancelled'::order_status"
  }
  check "orders_code_check" {
    expr = "code ~ '^CH-[0-9]{6}-[0-9]{3,}$'"
  }
  check "orders_subtotal_cents_check" {
    expr = "subtotal_cents >= 0"
  }
  check "orders_discount_cents_check" {
    expr = "discount_cents >= 0"
  }
  check "orders_total_cents_check" {
    expr = "total_cents >= 0"
  }
  check "orders_total_balance_check" {
    expr = "total_cents = subtotal_cents - discount_cents"
  }
  check "orders_terms_pair_check" {
    expr = "(terms_accepted_at IS NULL) = (terms_version IS NULL)"
  }
  // An order belongs to an account or to a guest with a name and a phone,
  // never both; only staff take a guest's order.
  check "orders_customer_check" {
    expr = "(user_id IS NULL) = (guest_name IS NOT NULL) AND (guest_name IS NULL) = (guest_phone IS NULL) AND (channel <> 'online' OR user_id IS NOT NULL)"
  }
}

table "order_items" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "order_id" {
    type = uuid
    null = false
  }
  column "line_type" {
    type = enum.line_type
    null = false
  }
  column "product_id" {
    type = uuid
    null = true
  }
  column "combo_id" {
    type = uuid
    null = true
  }
  column "configuration" {
    type    = jsonb
    null    = false
    default = sql("'{}'::jsonb")
  }
  column "name" {
    type = text
    null = false
  }
  column "slug" {
    type = text
    null = false
  }
  column "quantity" {
    type = int
    null = false
  }
  column "unit_price_cents" {
    type = bigint
    null = false
  }
  column "line_total_cents" {
    type = bigint
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "order_items_order_id_fkey" {
    columns     = [column.order_id]
    ref_columns = [table.orders.column.id]
    on_delete   = CASCADE
  }
  index "order_items_order_id_idx" {
    columns = [column.order_id]
  }
  check "order_items_quantity_check" {
    expr = "quantity > 0"
  }
  check "order_items_unit_price_cents_check" {
    expr = "unit_price_cents >= 0"
  }
  check "order_items_line_total_cents_check" {
    expr = "line_total_cents >= 0"
  }
  check "order_items_line_target_check" {
    expr = "(line_type = 'product' AND product_id IS NOT NULL AND combo_id IS NULL) OR (line_type = 'combo' AND combo_id IS NOT NULL AND product_id IS NULL)"
  }
}

// The part of an order one station makes. An order has at most one ticket per
// station; its status follows its tickets.
table "order_tickets" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "order_id" {
    type = uuid
    null = false
  }
  column "station" {
    type = enum.station
    null = false
  }
  column "status" {
    type    = enum.ticket_status
    null    = false
    default = sql("'queued'::ticket_status")
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "order_tickets_order_id_fkey" {
    columns     = [column.order_id]
    ref_columns = [table.orders.column.id]
    on_delete   = CASCADE
  }
  index "order_tickets_order_station_idx" {
    unique  = true
    columns = [column.order_id, column.station]
  }
  // A station's queue: its tickets still to make or collect.
  index "order_tickets_station_queue_idx" {
    columns = [column.station, column.status]
    where   = "status <> 'cancelled'::ticket_status"
  }
}

// What a ticket makes: products, a combo line counted by the products it holds.
// Names are kept as they were at checkout, like order_items.
table "order_ticket_items" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "ticket_id" {
    type = uuid
    null = false
  }
  column "order_item_id" {
    type = uuid
    null = false
  }
  column "product_id" {
    type = uuid
    null = false
  }
  column "name" {
    type = text
    null = false
  }
  column "quantity" {
    type = integer
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "order_ticket_items_ticket_id_fkey" {
    columns     = [column.ticket_id]
    ref_columns = [table.order_tickets.column.id]
    on_delete   = CASCADE
  }
  foreign_key "order_ticket_items_order_item_id_fkey" {
    columns     = [column.order_item_id]
    ref_columns = [table.order_items.column.id]
    on_delete   = CASCADE
  }
  index "order_ticket_items_ticket_id_idx" {
    columns = [column.ticket_id]
  }
  check "order_ticket_items_quantity_check" {
    expr = "quantity > 0"
  }
}

// The last order number handed out on each bakery day; checkout takes the next
// one under this row's lock, so numbers are unique and gap-free per day.
table "order_day_counters" {
  schema = schema.public
  column "day" {
    type = date
    null = false
  }
  column "last_number" {
    type = integer
    null = false
  }
  primary_key {
    columns = [column.day]
  }
  check "order_day_counters_last_number_check" {
    expr = "last_number > 0"
  }
}

// Every status an order has entered, when, and who moved it there. from_status
// is null for the row written when the order was placed.
table "order_status_events" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "order_id" {
    type = uuid
    null = false
  }
  column "from_status" {
    type = enum.order_status
    null = true
  }
  column "to_status" {
    type = enum.order_status
    null = false
  }
  // Both empty: the system made the move, as when an unpaid order expires.
  column "actor_id" {
    type = uuid
    null = true
  }
  column "actor_role" {
    type = enum.user_role
    null = true
  }
  // Why the bakery cancelled the order; the customer reads it.
  column "reason" {
    type = text
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "order_status_events_order_id_fkey" {
    columns     = [column.order_id]
    ref_columns = [table.orders.column.id]
    on_delete   = CASCADE
  }
  foreign_key "order_status_events_actor_id_fkey" {
    columns     = [column.actor_id]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  index "order_status_events_order_created_idx" {
    columns = [column.order_id, column.created_at]
  }
  // The moves to ready, found by when they happened: production time.
  index "order_status_events_ready_created_idx" {
    columns = [column.created_at]
    where   = "to_status = 'ready'::order_status"
  }
  check "order_status_events_move_check" {
    expr = "from_status IS DISTINCT FROM to_status"
  }
  check "order_status_events_actor_check" {
    expr = "(actor_id IS NULL) = (actor_role IS NULL)"
  }
  check "order_status_events_reason_check" {
    expr = "reason IS NULL OR (to_status = 'cancelled'::order_status AND char_length(reason) BETWEEN 1 AND 200)"
  }
}

// Transactional outbox: a row is written in the same transaction as the change
// it announces, so an event exists exactly when the change committed. Rows are
// technical delivery records, not business data: published rows are deleted
// after the retention window instead of being soft-deleted.
table "outbox_events" {
  schema = schema.public
  column "id" {
    type = uuid
    null = false
  }
  column "topic" {
    type = text
    null = false
  }
  column "audience" {
    type = jsonb
    null = false
  }
  column "data" {
    type = jsonb
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    // The database clock, so the API that writes a row and the worker that
    // sweeps it compare times from one source.
    default = sql("clock_timestamp()")
  }
  column "published_at" {
    type = timestamptz
    null = true
  }
  column "attempts" {
    type    = integer
    null    = false
    default = 0
  }
  column "last_error" {
    type = text
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  index "outbox_events_unpublished_idx" {
    columns = [column.created_at]
    where   = "(published_at IS NULL)"
  }
  index "outbox_events_published_idx" {
    columns = [column.published_at]
    where   = "(published_at IS NOT NULL)"
  }
}

// The bakery's pickup settings, edited by admins. Exactly one row (id = 1),
// seeded with the rules that were constants before.
table "store_settings" {
  schema = schema.public
  column "id" {
    type    = smallint
    null    = false
    default = 1
  }
  column "opens_at_minute" {
    type = smallint
    null = false
  }
  column "closes_at_minute" {
    type = smallint
    null = false
  }
  column "preorder_min_lead_minutes" {
    type = int
    null = false
  }
  column "max_advance_days" {
    type = smallint
    null = false
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "slot_minutes" {
    type    = smallint
    null    = false
    default = 30
  }
  column "slot_capacity" {
    type    = smallint
    null    = false
    default = 10
  }
  column "instant_prep_minutes" {
    type    = smallint
    null    = false
    default = 20
  }
  // How long an order placed online holds its slot and discount unpaid.
  column "payment_hold_minutes" {
    type    = smallint
    null    = false
    default = 15
  }
  primary_key {
    columns = [column.id]
  }
  check "store_settings_singleton_check" {
    expr = "id = 1"
  }
  check "store_settings_hours_check" {
    expr = "opens_at_minute >= 0 AND closes_at_minute < 1440 AND opens_at_minute < closes_at_minute"
  }
  check "store_settings_lead_check" {
    expr = "preorder_min_lead_minutes >= 0 AND preorder_min_lead_minutes <= 10080"
  }
  check "store_settings_advance_check" {
    expr = "max_advance_days >= 1 AND max_advance_days <= 90"
  }
  check "store_settings_slot_check" {
    expr = "slot_minutes IN (10, 15, 20, 30, 60) AND closes_at_minute - opens_at_minute >= slot_minutes"
  }
  check "store_settings_capacity_check" {
    expr = "slot_capacity >= 1 AND slot_capacity <= 200"
  }
  check "store_settings_instant_prep_check" {
    expr = "instant_prep_minutes >= 0 AND instant_prep_minutes <= 240"
  }
  check "store_settings_payment_hold_check" {
    expr = "payment_hold_minutes >= 5 AND payment_hold_minutes <= 120"
  }
}

// Days the bakery takes no pickups, with the reason customers are shown.
table "store_closed_dates" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "closed_on" {
    type = date
    null = false
  }
  column "reason" {
    type = text
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  index "store_closed_dates_day_idx" {
    unique  = true
    columns = [column.closed_on]
    where   = "(deleted_at IS NULL)"
  }
  check "store_closed_dates_reason_check" {
    expr = "char_length(reason) >= 1 AND char_length(reason) <= 200"
  }
}

enum "user_token_purpose" {
  schema = schema.public
  values = ["verify_email", "reset_password"]
}

// Single-use links emailed to a user: confirming their address, choosing a new
// password. Only a SHA-256 of the token is kept, so the table alone opens
// nothing. A user holds at most one token per purpose — issuing a new one
// replaces the last — and redeeming deletes it.
table "user_tokens" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "purpose" {
    type = enum.user_token_purpose
    null = false
  }
  column "token_hash" {
    type = bytea
    null = false
  }
  column "expires_at" {
    type = timestamptz
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "user_tokens_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = CASCADE
  }
  index "user_tokens_token_hash_idx" {
    unique  = true
    columns = [column.token_hash]
  }
  index "user_tokens_user_purpose_idx" {
    unique  = true
    columns = [column.user_id, column.purpose]
  }
  check "user_tokens_token_hash_check" {
    expr = "octet_length(token_hash) = 32"
  }
}

enum "payment_provider" {
  schema = schema.public
  values = ["paypal", "cash"]
}

// created: the buyer was sent to approve; pending: the provider holds the
// capture for review; captured: the money is taken; denied: the provider
// refused a pending capture, and the buyer may pay again.
enum "payment_status" {
  schema = schema.public
  values = ["created", "pending", "captured", "denied", "refunded"]
}

// How an order is paid: one row per order, never deleted. Only the provider's
// ids, the amount and the outcome are kept, never its payload, which carries
// the payer's personal details.
table "payments" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "order_id" {
    type = uuid
    null = false
  }
  column "provider" {
    type = enum.payment_provider
    null = false
  }
  // PayPal's order, and the page the buyer approves the payment on; cash has neither.
  column "provider_order_id" {
    type = text
    null = true
  }
  column "approve_url" {
    type = text
    null = true
  }
  column "status" {
    type    = enum.payment_status
    null    = false
    default = "created"
  }
  column "capture_id" {
    type = text
    null = true
  }
  column "amount_cents" {
    type = bigint
    null = false
  }
  column "currency" {
    type = text
    null = false
  }
  column "captured_at" {
    type = timestamptz
    null = true
  }
  // Set in the transaction that cancels the paid order; the refund follows.
  column "refund_requested_at" {
    type = timestamptz
    null = true
  }
  // Empty when PayPal reports the capture already refunded, as from its dashboard.
  column "refund_id" {
    type = text
    null = true
  }
  column "refunded_at" {
    type = timestamptz
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "payments_order_id_fkey" {
    columns     = [column.order_id]
    ref_columns = [table.orders.column.id]
    on_delete   = RESTRICT
  }
  index "payments_order_id_idx" {
    unique  = true
    columns = [column.order_id]
  }
  index "payments_provider_order_idx" {
    unique  = true
    columns = [column.provider, column.provider_order_id]
  }
  // Refunds the worker still has to make, oldest request first.
  index "payments_refund_due_idx" {
    columns = [column.refund_requested_at]
    where   = "status = 'captured'::payment_status AND refund_requested_at IS NOT NULL"
  }
  index "payments_capture_idx" {
    unique  = true
    columns = [column.provider, column.capture_id]
    where   = "capture_id IS NOT NULL"
  }
  check "payments_amount_cents_check" {
    expr = "amount_cents > 0"
  }
  check "payments_currency_check" {
    expr = "currency ~ '^[A-Z]{3}$'"
  }
  check "payments_capture_check" {
    expr = "provider <> 'paypal' OR status = 'denied' OR ((status = 'created') = (capture_id IS NULL))"
  }
  check "payments_provider_check" {
    expr = "(provider = 'paypal') = (provider_order_id IS NOT NULL AND approve_url IS NOT NULL)"
  }
  // Cash is due until the counter takes it at handoff: never captured by a
  // provider, never refunded, since it is taken only with the order.
  check "payments_cash_check" {
    expr = "provider <> 'cash' OR (capture_id IS NULL AND status IN ('created', 'captured') AND refund_requested_at IS NULL)"
  }
  check "payments_captured_at_check" {
    expr = "(status IN ('captured', 'refunded')) = (captured_at IS NOT NULL)"
  }
  check "payments_refunded_at_check" {
    expr = "(status = 'refunded') = (refunded_at IS NOT NULL)"
  }
  check "payments_refund_requested_check" {
    expr = "(refund_requested_at IS NULL OR captured_at IS NOT NULL) AND (refunded_at IS NULL OR refund_requested_at IS NOT NULL)"
  }
}

enum "conversation_status" {
  schema = schema.public
  values = ["open", "closed"]
}

// A customer and the counter writing to each other about one order. What each
// side has not read yet is counted on the row, so the inbox and the order list
// read it without counting messages. The counter shares one inbox: a message
// one staff member reads is read for all of them.
table "conversations" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "order_id" {
    type = uuid
    null = false
  }
  // Closed once the counter has resolved it; the next message opens it again.
  column "status" {
    type    = enum.conversation_status
    null    = false
    default = "open"
  }
  // The staff member who replied last; none until the counter answers.
  column "assigned_staff_id" {
    type = uuid
    null = true
  }
  column "unread_by_customer" {
    type    = integer
    null    = false
    default = 0
  }
  column "unread_by_staff" {
    type    = integer
    null    = false
    default = 0
  }
  column "last_message_at" {
    type = timestamptz
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "conversations_order_id_fkey" {
    columns     = [column.order_id]
    ref_columns = [table.orders.column.id]
    on_delete   = RESTRICT
  }
  foreign_key "conversations_assigned_staff_id_fkey" {
    columns     = [column.assigned_staff_id]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  index "conversations_order_id_idx" {
    unique  = true
    columns = [column.order_id]
  }
  // The inbox: latest message first, within a status.
  index "conversations_status_last_message_idx" {
    columns = [column.status, column.last_message_at]
  }
  check "conversations_unread_check" {
    expr = "unread_by_customer >= 0 AND unread_by_staff >= 0"
  }
}

// One message of a conversation, written by its customer or by staff. Erasing
// the customer's account erases every message: the row stays, its text goes.
table "messages" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "conversation_id" {
    type = uuid
    null = false
  }
  column "author_id" {
    type = uuid
    null = false
  }
  column "author_role" {
    type = enum.user_role
    null = false
  }
  column "body" {
    type = text
    null = false
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "messages_conversation_id_fkey" {
    columns     = [column.conversation_id]
    ref_columns = [table.conversations.column.id]
    on_delete   = RESTRICT
  }
  foreign_key "messages_author_id_fkey" {
    columns     = [column.author_id]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  // A thread reads newest first, a page at a time.
  index "messages_conversation_created_idx" {
    columns = [column.conversation_id, column.created_at, column.id]
  }
  check "messages_author_role_check" {
    expr = "author_role IN ('customer'::user_role, 'staff'::user_role)"
  }
  check "messages_body_check" {
    expr = "CASE WHEN deleted_at IS NULL THEN char_length(body) BETWEEN 1 AND 2000 ELSE body = '' END"
  }
}

enum "saved_list" {
  schema = schema.public
  values = ["favorite", "wishlist"]
}

// Products a customer keeps to come back to: favorites they order again, and
// a wishlist of what they want to try. One row per customer, product and list;
// removing a product marks the row, and saving it again brings the row back.
table "saved_products" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "product_id" {
    type = uuid
    null = false
  }
  column "list" {
    type = enum.saved_list
    null = false
  }
  // When it was last put on the list; the list shows the latest first.
  column "saved_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "saved_products_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  foreign_key "saved_products_product_id_fkey" {
    columns     = [column.product_id]
    ref_columns = [table.products.column.id]
    on_delete   = RESTRICT
  }
  index "saved_products_user_product_list_idx" {
    unique  = true
    columns = [column.user_id, column.product_id, column.list]
  }
  // A customer's lists, latest first.
  index "saved_products_user_list_saved_idx" {
    columns = [column.user_id, column.list, column.saved_at]
    where   = "(deleted_at IS NULL)"
  }
}

enum "review_status" {
  schema = schema.public
  values = ["pending", "published", "hidden"]
}

// A customer's review of a product they collected: one per product per order,
// held for a manager to publish or hide before anyone else reads it.
table "reviews" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "order_id" {
    type = uuid
    null = false
  }
  column "product_id" {
    type = uuid
    null = false
  }
  column "user_id" {
    type = uuid
    null = false
  }
  column "rating" {
    type = smallint
    null = false
  }
  // What the customer wrote, if anything; erasing their account clears it.
  column "comment" {
    type = text
    null = true
  }
  column "status" {
    type    = enum.review_status
    null    = false
    default = "pending"
  }
  // The manager who last published or hid it, and when.
  column "moderated_by" {
    type = uuid
    null = true
  }
  column "moderated_at" {
    type = timestamptz
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "reviews_order_id_fkey" {
    columns     = [column.order_id]
    ref_columns = [table.orders.column.id]
    on_delete   = RESTRICT
  }
  foreign_key "reviews_product_id_fkey" {
    columns     = [column.product_id]
    ref_columns = [table.products.column.id]
    on_delete   = RESTRICT
  }
  foreign_key "reviews_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  foreign_key "reviews_moderated_by_fkey" {
    columns     = [column.moderated_by]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  index "reviews_order_product_idx" {
    unique  = true
    columns = [column.order_id, column.product_id]
  }
  // A product's published reviews, latest first.
  index "reviews_product_published_idx" {
    columns = [column.product_id, column.created_at]
    where   = "((status = 'published'::review_status) AND (deleted_at IS NULL))"
  }
  // The moderation queue, by status, latest first.
  index "reviews_status_created_idx" {
    columns = [column.status, column.created_at]
    where   = "(deleted_at IS NULL)"
  }
  index "reviews_user_id_idx" {
    columns = [column.user_id]
  }
  check "reviews_rating_check" {
    expr = "rating BETWEEN 1 AND 5"
  }
  check "reviews_comment_check" {
    expr = "comment IS NULL OR char_length(comment) BETWEEN 1 AND 1000"
  }
  check "reviews_moderation_check" {
    expr = "(moderated_by IS NULL) = (moderated_at IS NULL) AND (status = 'pending') = (moderated_at IS NULL)"
  }
}

enum "promotion_status" {
  schema = schema.public
  values = ["sending", "sent"]
}

// A promotion a manager emailed to the customers who agreed to receive them:
// sending while the worker queues one email per customer, then sent to how many.
table "promotions" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "subject" {
    type = text
    null = false
  }
  column "body" {
    type = text
    null = false
  }
  column "status" {
    type    = enum.promotion_status
    null    = false
    default = "sending"
  }
  column "created_by" {
    type = uuid
    null = false
  }
  column "recipient_count" {
    type = integer
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  column "updated_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "promotions_created_by_fkey" {
    columns     = [column.created_by]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  // The promotions a manager sent, latest first.
  index "promotions_created_at_idx" {
    columns = [column.created_at]
  }
  check "promotions_subject_check" {
    expr = "char_length(subject) BETWEEN 1 AND 120"
  }
  check "promotions_body_check" {
    expr = "char_length(body) BETWEEN 1 AND 5000"
  }
  check "promotions_sent_check" {
    expr = "(status = 'sent') = (recipient_count IS NOT NULL)"
  }
  check "promotions_recipient_count_check" {
    expr = "recipient_count >= 0"
  }
}

enum "order_incident_type" {
  schema = schema.public
  values = ["bakery_cancelled", "ready_late", "no_show", "payment_failed", "payment_expired", "refunded", "payment_anomaly", "wrong_items", "custom_mismatch", "other"]
}

// Something that went wrong with an order: the system records the first seven
// types as they happen, staff report the last three.
table "order_incidents" {
  schema = schema.public
  column "id" {
    type    = uuid
    null    = false
    default = sql("gen_random_uuid()")
  }
  column "order_id" {
    type = uuid
    null = false
  }
  column "type" {
    type = enum.order_incident_type
    null = false
  }
  // What staff wrote reporting it; cleared when the customer erases their
  // account. A bakery cancellation's reason stays in the order's history.
  column "note" {
    type = text
    null = true
  }
  // Both empty: the system recorded it.
  column "actor_id" {
    type = uuid
    null = true
  }
  column "actor_role" {
    type = enum.user_role
    null = true
  }
  column "created_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "order_incidents_order_id_fkey" {
    columns     = [column.order_id]
    ref_columns = [table.orders.column.id]
    on_delete   = CASCADE
  }
  foreign_key "order_incidents_actor_id_fkey" {
    columns     = [column.actor_id]
    ref_columns = [table.users.column.id]
    on_delete   = RESTRICT
  }
  // The incidents of a week, latest first.
  index "order_incidents_created_at_idx" {
    columns = [column.created_at]
  }
  // An order's incidents, and a customer's recent ones through their orders.
  index "order_incidents_order_created_idx" {
    columns = [column.order_id, column.created_at]
  }
  check "order_incidents_actor_check" {
    expr = "(actor_id IS NULL) = (actor_role IS NULL)"
  }
  check "order_incidents_reported_check" {
    expr = "type NOT IN ('wrong_items', 'custom_mismatch', 'other') OR actor_id IS NOT NULL"
  }
  check "order_incidents_note_check" {
    expr = "note IS NULL OR (char_length(note) BETWEEN 1 AND 500 AND type IN ('wrong_items', 'custom_mismatch', 'other'))"
  }
}
