-- Initial schema for sqlc (keep aligned with db/schema.hcl and migrations/).
CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;

CREATE TYPE user_role AS ENUM ('admin', 'customer', 'staff', 'baker', 'manager');

CREATE TABLE users (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email                 citext NOT NULL,
    password_hash         text NOT NULL,
    role                  user_role NOT NULL DEFAULT 'customer',
    email_verified_at     timestamptz,
    must_change_password  boolean NOT NULL DEFAULT false,
    created_at            timestamptz NOT NULL DEFAULT now(),
    updated_at            timestamptz NOT NULL DEFAULT now(),
    deleted_at            timestamptz
);

CREATE UNIQUE INDEX users_email_active_idx ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX users_role_idx ON users (role) WHERE deleted_at IS NULL;

CREATE TABLE customer_profiles (
    user_id      uuid PRIMARY KEY,
    display_name text,
    phone        text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT customer_profiles_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE staff_profiles (
    user_id       uuid PRIMARY KEY,
    full_name     text NOT NULL DEFAULT '',
    phone         text,
    employee_code citext NOT NULL,
    hire_date     date NOT NULL,
    shift         text NOT NULL DEFAULT '',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT staff_profiles_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX staff_profiles_employee_code_idx ON staff_profiles (employee_code);

CREATE TABLE admin_profiles (
    user_id    uuid PRIMARY KEY,
    full_name  text NOT NULL DEFAULT '',
    phone      text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT admin_profiles_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE audit_logs (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id     uuid NOT NULL,
    actor_role   user_role NOT NULL,
    action       text NOT NULL,
    target_id    uuid,
    target_type  text NOT NULL,
    before_jsonb jsonb NOT NULL DEFAULT '{}'::jsonb,
    after_jsonb  jsonb NOT NULL DEFAULT '{}'::jsonb,
    ip           inet,
    user_agent   text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT audit_logs_actor_id_fkey
        FOREIGN KEY (actor_id) REFERENCES users (id) ON DELETE RESTRICT
);

CREATE INDEX audit_logs_actor_created_idx ON audit_logs (actor_id, created_at DESC);
CREATE INDEX audit_logs_target_idx ON audit_logs (target_type, target_id);
