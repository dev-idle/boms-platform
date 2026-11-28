import { z } from "zod";

import { orderCodeSchema, orderStatusSchema } from "@/lib/schemas/order";
import {
  stationSchema,
  ticketItemSchema,
  ticketStatusSchema,
  ticketSummarySchema,
  type PatchTicketStatusInput,
  type TicketItem,
  type TicketStatus,
} from "@/lib/schemas/ticket";
import { apiDateTimeSchema } from "@/lib/validation/datetime";

/** The statuses the kitchen's queue holds: a cancelled ticket leaves it. */
const activeTicketStatusSchema = ticketStatusSchema.exclude(["cancelled"]);

export type ActiveTicketStatus = z.infer<typeof activeTicketStatusSchema>;

/**
 * A ticket in the kitchen's queue. The kitchen gets a name for the order and
 * no way to contact the customer.
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
  /** One bakery day's pickups (YYYY-MM-DD); every day's when left out. */
  date: z.iso.date().optional(),
  /** Part of an order code, a customer's name or an item on the ticket. */
  search: z.string().trim().default(""),
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

/** A kitchen ticket with where the order's other tickets stand. */
export const kitchenTicketSchema = stationTicketSchema.extend({
  order_tickets: z.array(ticketSummarySchema),
});

export type KitchenTicket = z.infer<typeof kitchenTicketSchema>;

/** The kitchen's next move on a ticket in this status, and what it commits the bakery to; none once it is ready. */
export function nextTicketAction(status: TicketStatus): {
  label: string;
  title: string;
  consequence: string;
  status: PatchTicketStatusInput["status"];
} | null {
  switch (status) {
    case "queued":
      return {
        label: "Start",
        title: "Start this ticket?",
        consequence: "Once the kitchen starts, the customer can no longer cancel the order or move its pickup.",
        status: "in_progress",
      };
    case "in_progress":
      return {
        label: "Mark ready",
        title: "Mark this ticket ready?",
        consequence: "Once every station's part is ready, the customer is told the order can be collected.",
        status: "ready",
      };
    default:
      return null;
  }
}

/** How many products a ticket makes in all. */
export function ticketItemCount(items: ReadonlyArray<TicketItem>): number {
  return items.reduce((sum, item) => sum + item.quantity, 0);
}
