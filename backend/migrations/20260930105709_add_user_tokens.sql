-- Create enum type "user_token_purpose"
CREATE TYPE "user_token_purpose" AS ENUM ('verify_email', 'reset_password');
-- Create "user_tokens" table
CREATE TABLE "user_tokens" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "purpose" "user_token_purpose" NOT NULL,
  "token_hash" bytea NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "user_tokens_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "user_tokens_token_hash_check" CHECK (octet_length(token_hash) = 32)
);
-- Create index "user_tokens_token_hash_idx" to table: "user_tokens"
CREATE UNIQUE INDEX "user_tokens_token_hash_idx" ON "user_tokens" ("token_hash");
-- Create index "user_tokens_user_purpose_idx" to table: "user_tokens"
CREATE UNIQUE INDEX "user_tokens_user_purpose_idx" ON "user_tokens" ("user_id", "purpose");
-- Accounts from before email verification existed count as verified: they
-- signed up and ordered under the rules of their time, and asking every one of
-- them to confirm now would stop them checking out.
UPDATE "users" SET "email_verified_at" = "created_at" WHERE "email_verified_at" IS NULL;
