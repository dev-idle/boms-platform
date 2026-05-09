-- Baseline schema for sqlc. Replace or extend with your migration tool (goose, atlas, …).
CREATE TABLE IF NOT EXISTS schema_version (
    id         bigserial PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT now()
);
