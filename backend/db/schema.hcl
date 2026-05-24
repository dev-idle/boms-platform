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
  values = ["customer", "admin", "staff", "baker", "manager"]
}

table "schema_version" {
  schema = schema.public
  column "id" {
    type = bigint
    null = false
    identity {
      generated = BY_DEFAULT
      start     = 1
      increment = 1
    }
  }
  column "applied_at" {
    type    = timestamptz
    null    = false
    default = sql("now()")
  }
  primary_key {
    columns = [column.id]
  }
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
  column "hire_date" {
    type = date
    null = false
  }
  column "shift" {
    type    = text
    null    = false
    default = ""
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
