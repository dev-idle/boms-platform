-- Modify "orders" table
ALTER TABLE "orders" ADD CONSTRAINT "orders_terms_pair_check" CHECK ((terms_accepted_at IS NULL) = (terms_version IS NULL)), ADD COLUMN "terms_accepted_at" timestamptz NULL, ADD COLUMN "terms_version" text NULL;
-- Modify "users" table
ALTER TABLE "users" ADD CONSTRAINT "users_terms_pair_check" CHECK ((terms_accepted_at IS NULL) = (terms_version IS NULL)), ADD COLUMN "terms_accepted_at" timestamptz NULL, ADD COLUMN "terms_version" text NULL;
