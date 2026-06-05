CREATE TYPE discount_type AS ENUM ('percent', 'fixed_cents');

CREATE TABLE combos (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  slug citext NOT NULL,
  price_cents bigint NOT NULL,
  starts_at timestamptz NOT NULL,
  ends_at timestamptz NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT combos_price_cents_check CHECK (price_cents >= 0),
  CONSTRAINT combos_window_check CHECK (ends_at > starts_at)
);
CREATE UNIQUE INDEX combos_slug_active_idx ON combos (slug) WHERE (deleted_at IS NULL);
CREATE INDEX combos_active_window_idx ON combos (is_active, starts_at, ends_at) WHERE (deleted_at IS NULL);

CREATE TABLE combo_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  combo_id uuid NOT NULL REFERENCES combos (id) ON DELETE CASCADE,
  product_id uuid NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
  quantity integer NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT combo_items_quantity_check CHECK (quantity > 0),
  CONSTRAINT combo_items_combo_product_unique UNIQUE (combo_id, product_id)
);
CREATE INDEX combo_items_combo_id_idx ON combo_items (combo_id);
