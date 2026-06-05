CREATE TABLE discount_codes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code citext NOT NULL,
  discount_type discount_type NOT NULL,
  value bigint NOT NULL,
  min_order_cents bigint,
  max_uses integer,
  used_count integer NOT NULL DEFAULT 0,
  starts_at timestamptz NOT NULL,
  ends_at timestamptz NOT NULL,
  is_active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT discount_codes_window_check CHECK (ends_at > starts_at),
  CONSTRAINT discount_codes_used_count_check CHECK (used_count >= 0),
  CONSTRAINT discount_codes_min_order_cents_check CHECK (min_order_cents IS NULL OR min_order_cents >= 0),
  CONSTRAINT discount_codes_max_uses_check CHECK (max_uses IS NULL OR max_uses > 0)
);
CREATE UNIQUE INDEX discount_codes_code_active_idx ON discount_codes (code) WHERE (deleted_at IS NULL);
CREATE INDEX discount_codes_active_window_idx ON discount_codes (is_active, starts_at, ends_at) WHERE (deleted_at IS NULL);
ALTER TABLE discount_codes
  ADD CONSTRAINT discount_codes_value_percent_check
    CHECK (discount_type <> 'percent' OR (value >= 1 AND value <= 100)),
  ADD CONSTRAINT discount_codes_value_fixed_cents_check
    CHECK (discount_type <> 'fixed_cents' OR value >= 1),
  ADD CONSTRAINT discount_codes_used_within_max_check
    CHECK (max_uses IS NULL OR used_count <= max_uses);
