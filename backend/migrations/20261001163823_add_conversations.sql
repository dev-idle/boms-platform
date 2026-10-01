-- Create enum type "conversation_status"
CREATE TYPE "conversation_status" AS ENUM ('open', 'closed');
-- Create "conversations" table
CREATE TABLE "conversations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "status" "conversation_status" NOT NULL DEFAULT 'open',
  "assigned_staff_id" uuid NULL,
  "unread_by_customer" integer NOT NULL DEFAULT 0,
  "unread_by_staff" integer NOT NULL DEFAULT 0,
  "last_message_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "conversations_assigned_staff_id_fkey" FOREIGN KEY ("assigned_staff_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "conversations_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "conversations_unread_check" CHECK ((unread_by_customer >= 0) AND (unread_by_staff >= 0))
);
-- Create index "conversations_order_id_idx" to table: "conversations"
CREATE UNIQUE INDEX "conversations_order_id_idx" ON "conversations" ("order_id");
-- Create index "conversations_status_last_message_idx" to table: "conversations"
CREATE INDEX "conversations_status_last_message_idx" ON "conversations" ("status", "last_message_at");
-- Create "messages" table
CREATE TABLE "messages" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "conversation_id" uuid NOT NULL,
  "author_id" uuid NOT NULL,
  "author_role" "user_role" NOT NULL,
  "body" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "messages_author_id_fkey" FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "messages_conversation_id_fkey" FOREIGN KEY ("conversation_id") REFERENCES "conversations" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "messages_author_role_check" CHECK (author_role = ANY (ARRAY['customer'::user_role, 'staff'::user_role])),
  CONSTRAINT "messages_body_check" CHECK (
CASE
    WHEN (deleted_at IS NULL) THEN ((char_length(body) >= 1) AND (char_length(body) <= 2000))
    ELSE (body = ''::text)
END)
);
-- Create index "messages_conversation_created_idx" to table: "messages"
CREATE INDEX "messages_conversation_created_idx" ON "messages" ("conversation_id", "created_at", "id");
