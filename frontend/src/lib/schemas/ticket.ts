import { z } from "zod";

import { customizationSchema, orderCodeSchema, orderStatusSchema } from "@/lib/schemas/order";
import { apiDateTimeSchema } from "@/lib/validation/datetime";

/** Where a category's products are made — backend `category.Station`. */
export const stationSchema = z.enum(["kitchen", "counter"]);

export type Station = z.infer<typeof stationSchema>;

/** How each station reads wherever it is named. */
export const STATION_LABEL: Record<Station, string> = {
  kitchen: "Kitchen",
  counter: "Counter",
};

/** A station's part of an order — mirrors backend `domain/order.TicketStatus`. */
export const ticketStatusSchema = z.enum(["queued", "in_progress", "ready", "cancelled"]);

export type TicketStatus = z.infer<typeof ticketStatusSchema>;

/** The statuses a station queue holds: a cancelled ticket leaves it. */
const activeTicketStatusSchema = ticketStatusSchema.exclude(["cancelled"]);

export type ActiveTicketStatus = z.infer<typeof activeTicketStatusSchema>;

/** Where one station's part of an order stands. */
export const ticketSummarySchema = z.object({
  station: stationSchema,
  status: ticketStatusSchema,
});

/** One product a ticket makes; a combo arrives as its component products. */
export const ticketItemSchema = z.object({
  name: z.string().min(1),
  quantity: z.number().int().min(1),
  /** How the customer configured it; null for a plain item and in a station's list. */
  customization: customizationSchema.nullable(),
});

export type TicketItem = z.infer<typeof ticketItemSchema>;

/**
 * A ticket in a station's queue. The station gets a name for the order and no
 * way to contact the customer.
 */
export const stationTicketSchema = z.object({
  id: z.uuid(),
  order_id: z.uuid(),
  order_code: orderCodeSchema,
  order_status: orderStatusSchema.extract(["confirmed", "in_production", "ready"]),
  station: stationSchema,
  status: activeTicketStatusSchema,
  pickup_at: apiDateTimeSchema.nullable().optional(),
  customer: z.object({ display_name: z.string().optional() }),
  items: z.array(ticketItemSchema),
  created_at: apiDateTimeSchema,
});

export type StationTicket = z.infer<typeof stationTicketSchema>;

export const stationTicketsListFilterSchema = z.object({
  page: z.number().int().min(1).default(1),
  page_size: z.number().int().min(1).max(100).default(20),
  status: activeTicketStatusSchema.optional(),
});

export type StationTicketsListFilterInput = z.infer<typeof stationTicketsListFilterSchema>;

export type StationTicketsListResult = {
  tickets: StationTicket[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
  };
  request_id?: string;
};

/** The move a station makes on its own ticket. */
export const patchTicketStatusInputSchema = z.object({
  status: z.enum(["in_progress", "ready"]),
});

export type PatchTicketStatusInput = z.infer<typeof patchTicketStatusInputSchema>;

/** A ticket after a move, with the status its order reached. */
export const ticketChangeSchema = z.object({
  id: z.uuid(),
  order_id: z.uuid(),
  station: stationSchema,
  status: ticketStatusSchema,
  order_status: orderStatusSchema,
});

export type TicketChange = z.infer<typeof ticketChangeSchema>;

/** The one move a station can make on a ticket in this status, if any. */
export function nextTicketAction(
  status: TicketStatus,
): { label: string; status: PatchTicketStatusInput["status"] } | null {
  switch (status) {
    case "queued":
      return { label: "Start", status: "in_progress" };
    case "in_progress":
      return { label: "Mark ready", status: "ready" };
    default:
      return null;
  }
}

/** A ticket's products on one line: "2× Matcha cake · 1× Croissant". */
export function formatTicketItems(items: ReadonlyArray<TicketItem>): string {
  return items.map((item) => `${item.quantity}× ${item.name}`).join(" · ");
}

/** How many products a ticket makes in all. */
export function ticketItemCount(items: ReadonlyArray<TicketItem>): number {
  return items.reduce((sum, item) => sum + item.quantity, 0);
}

/** What a station hears after moving a ticket, including when that finished the order. */
export function ticketChangeMessage(change: TicketChange): string {
  if (change.status === "in_progress") {
    return "Ticket started";
  }
  return change.order_status === "ready"
    ? "Ticket ready — the whole order is ready for pickup"
    : "Ticket ready";
}
