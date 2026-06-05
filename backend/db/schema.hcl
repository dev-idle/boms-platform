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
  primary_key {
    columns = [column.id]
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
  column "is_available" {
    type    = boolean
    null    = false
    default = true
  }
  column "image_url" {
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
  column "deleted_at" {
    type = timestamptz
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
  check "discount_codes_value_percent_check" {
    expr = "discount_type <> 'percent' OR (value >= 1 AND value <= 100)"
  }
  check "discount_codes_value_fixed_cents_check" {
    expr = "discount_type <> 'fixed_cents' OR value >= 1"
  }
  check "discount_codes_used_within_max_check" {
    expr = "max_uses IS NULL OR used_count <= max_uses"
  }
}

enum "cart_line_type" {
  schema = schema.public
  values = ["product", "combo"]
}

enum "order_status" {
  schema = schema.public
  values = ["pending", "confirmed", "cancelled", "fulfilled"]
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
    type = enum.cart_line_type
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
  index "cart_items_cart_product_idx" {
    unique  = true
    columns = [column.cart_id, column.product_id]
    where   = "line_type = 'product'"
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
  column "user_id" {
    type = uuid
    null = false
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
    type = enum.cart_line_type
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
