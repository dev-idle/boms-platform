-- Catalog tables: categories and products (manager-owned, customer read-only via /catalog/*).

CREATE TABLE categories (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    slug       citext NOT NULL,
    sort_order int NOT NULL DEFAULT 0,
    is_active  boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE UNIQUE INDEX categories_slug_active_idx ON categories (slug) WHERE deleted_at IS NULL;
CREATE INDEX categories_active_sort_idx ON categories (is_active, sort_order) WHERE deleted_at IS NULL;

CREATE TABLE products (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id  uuid NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    name         text NOT NULL,
    slug         citext NOT NULL,
    description  text,
    price_cents  bigint NOT NULL CHECK (price_cents >= 0),
    is_available boolean NOT NULL DEFAULT true,
    image_url    text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

CREATE UNIQUE INDEX products_slug_active_idx ON products (slug) WHERE deleted_at IS NULL;
CREATE INDEX products_category_active_idx ON products (category_id) WHERE deleted_at IS NULL;
