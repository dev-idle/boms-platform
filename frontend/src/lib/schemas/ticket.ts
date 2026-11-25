import { z } from "zod";

import { customizationSchema, orderStatusSchema } from "@/lib/schemas/order";

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

/** A ticket's products on one line: "2× Matcha cake · 1× Croissant". */
export function formatTicketItems(items: ReadonlyArray<TicketItem>): string {
  return items.map((item) => `${item.quantity}× ${item.name}`).join(" · ");
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
