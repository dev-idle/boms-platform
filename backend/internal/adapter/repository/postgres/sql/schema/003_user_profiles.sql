-- User module schema extensions. Keep aligned with db/schema.hcl and migrations/.
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'staff';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'baker';
ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'manager';

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS must_change_password boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS customer_profiles (
    user_id      uuid PRIMARY KEY,
    display_name text,
    phone        text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT customer_profiles_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS staff_profiles (
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

CREATE UNIQUE INDEX IF NOT EXISTS staff_profiles_employee_code_idx ON staff_profiles (employee_code);

CREATE TABLE IF NOT EXISTS admin_profiles (
    user_id    uuid PRIMARY KEY,
    full_name  text NOT NULL DEFAULT '',
    phone      text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT admin_profiles_user_id_fkey
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id    uuid NOT NULL,
    actor_role  user_role NOT NULL,
    action      text NOT NULL,
    target_id   uuid,
    target_type text NOT NULL,
    before_jsonb jsonb NOT NULL DEFAULT '{}'::jsonb,
    after_jsonb  jsonb NOT NULL DEFAULT '{}'::jsonb,
    ip          inet,
    user_agent  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT audit_logs_actor_id_fkey
        FOREIGN KEY (actor_id) REFERENCES users (id) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS audit_logs_actor_created_idx ON audit_logs (actor_id, created_at DESC);
CREATE INDEX IF NOT EXISTS audit_logs_target_idx ON audit_logs (target_type, target_id);
