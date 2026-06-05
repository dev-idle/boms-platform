-- Tighten discount_codes value and usage invariants at the database layer.

ALTER TABLE "discount_codes"
  ADD CONSTRAINT "discount_codes_value_percent_check"
    CHECK (discount_type <> 'percent' OR (value >= 1 AND value <= 100)),
  ADD CONSTRAINT "discount_codes_value_fixed_cents_check"
    CHECK (discount_type <> 'fixed_cents' OR value >= 1),
  ADD CONSTRAINT "discount_codes_used_within_max_check"
    CHECK (max_uses IS NULL OR used_count <= max_uses);
