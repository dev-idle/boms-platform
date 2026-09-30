import type { OrderStatus, OrderTimelineEntry } from "@/lib/schemas/order";

export const ORDER_PROGRESS_STEPS = [
  { key: "awaiting_payment", label: "Placed" },
  { key: "confirmed", label: "Confirmed" },
  { key: "in_production", label: "In production" },
  { key: "ready", label: "Ready" },
  { key: "fulfilled", label: "Picked up" },
] as const satisfies ReadonlyArray<{ key: OrderStatus; label: string }>;

export function activeOrderProgressIndex(status: OrderStatus): number {
  if (isOrderDropped(status)) {
    return -1;
  }
  // An order placed before online payment waited for the counter instead: it
  // is placed as well.
  if (status === "pending") {
    return 0;
  }
  return ORDER_PROGRESS_STEPS.findIndex((step) => step.key === status);
}

/** Whether the bakery works on the order, so its stations' progress means something. */
export function isWithTheBakery(status: OrderStatus): boolean {
  return status !== "awaiting_payment" && !isOrderDropped(status);
}

/** Whether the order will not be made: cancelled, or not paid in time. */
export function isOrderDropped(status: OrderStatus): boolean {
  return status === "cancelled" || status === "expired";
}

/** When the order entered status, from its timeline; undefined while it has not. */
export function statusReachedAt(
  timeline: ReadonlyArray<OrderTimelineEntry>,
  status: OrderStatus,
): string | undefined {
  return timeline.find((entry) => entry.status === status)?.at;
}
