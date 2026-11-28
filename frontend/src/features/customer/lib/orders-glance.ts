import type { OrderStatus } from "@/lib/schemas/order";
import { bakeryDayOf } from "@/lib/validation/pickup";

import type { OrderSummary } from "../schemas";

/** Orders the bakery has taken and not handed over yet. */
const TO_COLLECT: ReadonlySet<OrderStatus> = new Set(["pending", "confirmed", "in_production", "ready"]);

/**
 * The order still to collect with the earliest pickup from today on (bakery
 * time); null when there is none. A pickup earlier today still counts — the
 * order may be waiting at the counter — but one from a past day does not.
 */
export function nextPickup(orders: readonly OrderSummary[], now: Date): OrderSummary | null {
  const today = bakeryDayOf(now);
  let next: OrderSummary | null = null;
  for (const order of orders) {
    if (!TO_COLLECT.has(order.status) || !order.pickup_at) {
      continue;
    }
    const at = new Date(order.pickup_at);
    if (bakeryDayOf(at) < today) {
      continue;
    }
    if (!next?.pickup_at || at.getTime() < Date.parse(next.pickup_at)) {
      next = order;
    }
  }
  return next;
}

/** How many of the counter's messages the customer has not read, across these orders. */
export function unreadOrderMessages(orders: readonly OrderSummary[]): number {
  return orders.reduce((total, order) => total + order.unread_messages, 0);
}
