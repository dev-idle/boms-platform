import type { OrderStatus } from "../schemas";

export const ORDER_PROGRESS_STEPS = [
  { key: "pending", label: "Placed" },
  { key: "confirmed", label: "Confirmed" },
  { key: "in_production", label: "In production" },
  { key: "ready", label: "Ready" },
  { key: "fulfilled", label: "Picked up" },
] as const satisfies ReadonlyArray<{ key: OrderStatus; label: string }>;

export type OrderProgressStepKey = (typeof ORDER_PROGRESS_STEPS)[number]["key"];

export function activeOrderProgressIndex(status: OrderStatus): number {
  if (status === "cancelled") {
    return -1;
  }
  return ORDER_PROGRESS_STEPS.findIndex((step) => step.key === status);
}

export function isOrderCancelled(status: OrderStatus): boolean {
  return status === "cancelled";
}
