-- One transaction: the backfill rewrites every order row anyway, so a
-- concurrent index build would save no lock time and would split the backfill
-- from the constraint that guards it.

-- Create "order_day_counters" table
CREATE TABLE "order_day_counters" (
  "day" date NOT NULL,
  "last_number" integer NOT NULL,
  PRIMARY KEY ("day"),
  CONSTRAINT "order_day_counters_last_number_check" CHECK (last_number > 0)
);
-- Number the orders placed so far per bakery day (Asia/Ho_Chi_Minh), in the
-- order they were placed, and carry each day's count into its counter.
ALTER TABLE "orders" ADD COLUMN "code" text NULL;
UPDATE "orders" AS o
SET "code" = 'CH-' || to_char(n.day, 'YYMMDD') || '-' || lpad(n.number::text, greatest(3, length(n.number::text)), '0')
FROM (
  SELECT
    "id",
    ("created_at" AT TIME ZONE 'Asia/Ho_Chi_Minh')::date AS day,
    row_number() OVER (
      PARTITION BY ("created_at" AT TIME ZONE 'Asia/Ho_Chi_Minh')::date
      ORDER BY "created_at", "id"
    ) AS number
  FROM "orders"
) AS n
WHERE n.id = o.id;
INSERT INTO "order_day_counters" ("day", "last_number")
SELECT ("created_at" AT TIME ZONE 'Asia/Ho_Chi_Minh')::date, count(*)
FROM "orders"
GROUP BY 1;
-- Modify "orders" table
ALTER TABLE "orders" ALTER COLUMN "code" SET NOT NULL, ADD CONSTRAINT "orders_code_check" CHECK (code ~ '^CH-[0-9]{6}-[0-9]{3,}$'::text);
-- Create index "orders_code_idx" to table: "orders"
CREATE UNIQUE INDEX "orders_code_idx" ON "orders" ("code");
-- Create "order_status_events" table
CREATE TABLE "order_status_events" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "from_status" "order_status" NULL,
  "to_status" "order_status" NOT NULL,
  "actor_id" uuid NOT NULL,
  "actor_role" "user_role" NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_status_events_actor_id_fkey" FOREIGN KEY ("actor_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "order_status_events_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "order_status_events_move_check" CHECK (from_status IS DISTINCT FROM to_status)
);
-- Create index "order_status_events_order_created_idx" to table: "order_status_events"
CREATE INDEX "order_status_events_order_created_idx" ON "order_status_events" ("order_id", "created_at");
-- The history so far: every order was placed as pending by its customer (the
-- role checkout requires, whatever role the account holds now); the moves
-- staff and bakers made since are in the audit log.
INSERT INTO "order_status_events" ("order_id", "from_status", "to_status", "actor_id", "actor_role", "created_at")
SELECT "id", NULL, 'pending', "user_id", 'customer', "created_at"
FROM "orders";
INSERT INTO "order_status_events" ("order_id", "from_status", "to_status", "actor_id", "actor_role", "created_at")
SELECT a."target_id", (a."before_jsonb" ->> 'status')::order_status, (a."after_jsonb" ->> 'status')::order_status, a."actor_id", a."actor_role", a."created_at"
FROM "audit_logs" AS a
JOIN "orders" AS o ON o."id" = a."target_id"
WHERE a."target_type" = 'order'
  AND a."action" IN ('staff.updated_order_status', 'baker.updated_order_status')
  AND a."before_jsonb" ->> 'status' IS DISTINCT FROM a."after_jsonb" ->> 'status';
