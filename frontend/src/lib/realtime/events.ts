import { z } from "zod";

/** Event types the API pushes — mirrors the topics in the backend domain `event.go` files. */
export const REALTIME_EVENT_TYPE = {
  conversationChanged: "conversation.changed",
  messageCreated: "message.created",
  orderCreated: "order.created",
  orderStatusChanged: "order.status_changed",
  orderRescheduled: "order.rescheduled",
  orderRefunded: "order.refunded",
  productSoldOutChanged: "product.sold_out_changed",
  settingsUpdated: "settings.updated",
  slotsChanged: "slots.changed",
  ticketChanged: "ticket.changed",
} as const;

/**
 * A change notice: which change happened and where to look again, never the
 * changed record. Pages refetch through the API, where authorization lives.
 */
export const realtimeEventSchema = z.object({
  type: z.string().min(1),
  data: z.record(z.string(), z.string()),
});

export type RealtimeEvent = z.infer<typeof realtimeEventSchema>;

/** `POST /api/v1/realtime/tickets` — a single-use ticket and where to redeem it. */
export const realtimeTicketSchema = z.object({
  ticket: z.string().min(1),
  url: z.url({ protocol: /^wss?$/ }),
});

export type RealtimeTicket = z.infer<typeof realtimeTicketSchema>;
