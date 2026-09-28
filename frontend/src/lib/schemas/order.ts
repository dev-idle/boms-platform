import { z } from "zod";

import { apiDateTimeSchema } from "@/lib/validation/datetime";

/** Order lifecycle — mirrors backend `domain/order.Status`; the single frontend source. */
export const orderStatusSchema = z.enum([
  "pending",
  "confirmed",
  "in_production",
  "ready",
  "cancelled",
  "fulfilled",
]);

export type OrderStatus = z.infer<typeof orderStatusSchema>;

/** CH-YYMMDD-NNN: the bakery day an order was placed and its number that day (backend `order.Code`). */
export const orderCodeSchema = z.string().regex(/^CH-\d{6}-\d{3,}$/);

/** A status an order entered and when — one entry of `timeline` on an order detail. */
export const orderTimelineEntrySchema = z.object({
  status: orderStatusSchema,
  at: apiDateTimeSchema,
});

export type OrderTimelineEntry = z.infer<typeof orderTimelineEntrySchema>;

/**
 * Whether this status is the one being applied right now. A queue offers several
 * transitions at once, so only the button that was pressed may say it is working
 * — the rest are simply held.
 */
export function isApplyingOrderStatus(
  mutation: { isPending: boolean; variables?: { status: OrderStatus } },
  status: OrderStatus,
): boolean {
  return mutation.isPending && mutation.variables?.status === status;
}
