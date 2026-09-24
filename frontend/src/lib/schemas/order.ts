import { z } from "zod";

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
