-- User module: operational roles, profile tables, audit logs.
ALTER TYPE "user_role" ADD VALUE IF NOT EXISTS 'staff';
ALTER TYPE "user_role" ADD VALUE IF NOT EXISTS 'baker';
ALTER TYPE "user_role" ADD VALUE IF NOT EXISTS 'manager';

ALTER TABLE "users" ADD COLUMN "must_change_password" boolean NOT NULL DEFAULT false;

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
  "hire_date" date NOT NULL,
  "shift" text NOT NULL DEFAULT '',
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
