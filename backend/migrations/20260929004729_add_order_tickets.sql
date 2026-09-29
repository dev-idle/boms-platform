-- Create enum type "ticket_status"
CREATE TYPE "ticket_status" AS ENUM ('queued', 'in_progress', 'ready', 'cancelled');
-- Create "order_tickets" table
CREATE TABLE "order_tickets" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "order_id" uuid NOT NULL,
  "station" "station" NOT NULL,
  "status" "ticket_status" NOT NULL DEFAULT 'queued',
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_tickets_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "order_tickets_order_station_idx" to table: "order_tickets"
CREATE UNIQUE INDEX "order_tickets_order_station_idx" ON "order_tickets" ("order_id", "station");
-- Create index "order_tickets_station_queue_idx" to table: "order_tickets"
CREATE INDEX "order_tickets_station_queue_idx" ON "order_tickets" ("station", "status") WHERE (status <> 'cancelled'::ticket_status);
-- Create "order_ticket_items" table
CREATE TABLE "order_ticket_items" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "ticket_id" uuid NOT NULL,
  "order_item_id" uuid NOT NULL,
  "product_id" uuid NOT NULL,
  "name" text NOT NULL,
  "quantity" integer NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "order_ticket_items_order_item_id_fkey" FOREIGN KEY ("order_item_id") REFERENCES "order_items" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "order_ticket_items_ticket_id_fkey" FOREIGN KEY ("ticket_id") REFERENCES "order_tickets" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "order_ticket_items_quantity_check" CHECK (quantity > 0)
);
-- Create index "order_ticket_items_ticket_id_idx" to table: "order_ticket_items"
CREATE INDEX "order_ticket_items_ticket_id_idx" ON "order_ticket_items" ("ticket_id");
-- Tickets for the orders still in play (pending to ready): each product line,
-- and each product a combo line holds, goes to its category's station. The
-- ticket's status follows the order's: in production means started, ready
-- means ready, anything earlier means queued. Finished and cancelled orders
-- get none.
INSERT INTO "order_tickets" ("order_id", "station", "status", "created_at", "updated_at")
SELECT DISTINCT
  lines.order_id,
  lines.station,
  (CASE o."status"
    WHEN 'in_production' THEN 'in_progress'
    WHEN 'ready' THEN 'ready'
    ELSE 'queued'
  END)::ticket_status,
  o."created_at",
  o."updated_at"
FROM (
  SELECT oi."order_id", c."station"
  FROM "order_items" AS oi
  JOIN "products" AS p ON p."id" = oi."product_id"
  JOIN "categories" AS c ON c."id" = p."category_id"
  WHERE oi."line_type" = 'product'
  UNION
  SELECT oi."order_id", c."station"
  FROM "order_items" AS oi
  JOIN "combo_items" AS ci ON ci."combo_id" = oi."combo_id"
  JOIN "products" AS p ON p."id" = ci."product_id"
  JOIN "categories" AS c ON c."id" = p."category_id"
  WHERE oi."line_type" = 'combo'
) AS lines
JOIN "orders" AS o ON o."id" = lines.order_id
WHERE o."status" IN ('pending', 'confirmed', 'in_production', 'ready');
INSERT INTO "order_ticket_items" ("ticket_id", "order_item_id", "product_id", "name", "quantity", "created_at")
SELECT t."id", lines.order_item_id, lines.product_id, lines.name, lines.quantity, lines.created_at
FROM (
  SELECT oi."order_id", oi."id" AS order_item_id, p."id" AS product_id, oi."name", c."station", oi."quantity", oi."created_at"
  FROM "order_items" AS oi
  JOIN "products" AS p ON p."id" = oi."product_id"
  JOIN "categories" AS c ON c."id" = p."category_id"
  WHERE oi."line_type" = 'product'
  UNION ALL
  SELECT oi."order_id", oi."id", p."id", p."name", c."station", oi."quantity" * ci."quantity", oi."created_at"
  FROM "order_items" AS oi
  JOIN "combo_items" AS ci ON ci."combo_id" = oi."combo_id"
  JOIN "products" AS p ON p."id" = ci."product_id"
  JOIN "categories" AS c ON c."id" = p."category_id"
  WHERE oi."line_type" = 'combo'
) AS lines
JOIN "order_tickets" AS t ON t."order_id" = lines.order_id AND t."station" = lines.station;
