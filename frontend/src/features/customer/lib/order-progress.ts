import type { OrderStatus, OrderTimelineEntry } from "@/lib/schemas/order";

export const ORDER_PROGRESS_STEPS = [
  { key: "awaiting_payment", label: "Placed" },
  { key: "confirmed", label: "Confirmed" },
  { key: "in_production", label: "In production" },
  { key: "ready", label: "Ready" },
  { key: "fulfilled", label: "Picked up" },
] as const satisfies ReadonlyArray<{ key: OrderStatus; label: string }>;

export function activeOrderProgressIndex(status: OrderStatus): number {
  if (isUncollected(status)) {
    return -1;
  }
  // A paid order whose reference photo staff check first is placed as well; it
  // moves on once they accept it.
  if (status === "pending") {
    return 0;
  }
  return ORDER_PROGRESS_STEPS.findIndex((step) => step.key === status);
}

/** Whether the bakery works on the order, so its stations' progress means something. */
export function isWithTheBakery(status: OrderStatus): boolean {
  return status !== "awaiting_payment" && !isUncollected(status);
}

/** Whether the order ended without being collected: cancelled, not paid in time, or not picked up. */
export function isUncollected(status: OrderStatus): boolean {
  return status === "cancelled" || status === "expired" || status === "no_show";
}

/** When the order entered status, from its timeline; undefined while it has not. */
export function statusReachedAt(
  timeline: ReadonlyArray<OrderTimelineEntry>,
  status: OrderStatus,
): string | undefined {
  return timeline.find((entry) => entry.status === status)?.at;
}
