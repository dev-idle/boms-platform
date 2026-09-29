import { z } from "zod";

import { stationTicketSchema, ticketSummarySchema } from "@/lib/schemas/ticket";

/** A kitchen ticket with where the order's other tickets stand. */
export const kitchenTicketSchema = stationTicketSchema.extend({
  order_tickets: z.array(ticketSummarySchema),
});

export type KitchenTicket = z.infer<typeof kitchenTicketSchema>;
