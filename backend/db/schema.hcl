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
  values = ["customer", "admin"]
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
