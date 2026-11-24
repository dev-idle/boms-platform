-- Create enum type "assistant_message_role"
CREATE TYPE "assistant_message_role" AS ENUM ('visitor', 'assistant');
-- Create "assistant_sessions" table
CREATE TABLE "assistant_sessions" (
  "id" uuid NOT NULL,
  "user_id" uuid NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "last_message_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "assistant_sessions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- Create index "assistant_sessions_last_message_idx" to table: "assistant_sessions"
CREATE INDEX "assistant_sessions_last_message_idx" ON "assistant_sessions" ("last_message_at");
-- Create index "assistant_sessions_user_idx" to table: "assistant_sessions"
CREATE INDEX "assistant_sessions_user_idx" ON "assistant_sessions" ("user_id") WHERE (user_id IS NOT NULL);
-- Create "assistant_messages" table
CREATE TABLE "assistant_messages" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "session_id" uuid NOT NULL,
  "role" "assistant_message_role" NOT NULL,
  "body" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "assistant_messages_session_id_fkey" FOREIGN KEY ("session_id") REFERENCES "assistant_sessions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "assistant_messages_body_check" CHECK ((char_length(body) >= 1) AND (char_length(body) <= 2000))
);
-- Create index "assistant_messages_session_created_idx" to table: "assistant_messages"
CREATE INDEX "assistant_messages_session_created_idx" ON "assistant_messages" ("session_id", "created_at", "id");
-- Modify "conversations" table: a conversation names its customer, so one
-- about no order has one; those already written take their order's customer.
ALTER TABLE "conversations" ADD COLUMN "customer_id" uuid NULL;
UPDATE "conversations" c SET "customer_id" = o."user_id" FROM "orders" o WHERE o."id" = c."order_id";
ALTER TABLE "conversations" ALTER COLUMN "customer_id" SET NOT NULL, ALTER COLUMN "order_id" DROP NOT NULL, ADD CONSTRAINT "conversations_customer_id_fkey" FOREIGN KEY ("customer_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT;
