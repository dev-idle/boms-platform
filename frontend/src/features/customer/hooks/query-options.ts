import type { QueryKey } from "@tanstack/react-query";

import { REALTIME_EVENT_TYPE, type RealtimeEvent } from "@/lib/realtime/events";

import type { OrdersListFilterInput } from "../schemas";

export const customerQueryKeys = {
  cart: ["customer", "cart"] as const,
  pickupRules: ["customer", "pickup-rules"] as const,
  ordersRoot: ["customer", "orders"] as const,
  orders: (filter: OrdersListFilterInput) =>
    [...customerQueryKeys.ordersRoot, filter] as const,
  orderRoot: ["customer", "order"] as const,
  order: (id: string) => [...customerQueryKeys.orderRoot, id] as const,
};

/** Everything pushed events can change in a customer tab, refetched after a gap. */
export const customerLiveQueryKeys: readonly QueryKey[] = [
  customerQueryKeys.cart,
  customerQueryKeys.pickupRules,
  customerQueryKeys.ordersRoot,
  customerQueryKeys.orderRoot,
];

/**
 * Queries a pushed event makes stale in a customer tab. Checkout empties the
 * cart, so a new order placed in another tab refreshes the cart here too; a
 * settings change refreshes the pickup window the cart offers.
 */
export function customerQueryKeysForEvent(event: RealtimeEvent): QueryKey[] {
  switch (event.type) {
    case REALTIME_EVENT_TYPE.orderCreated:
      return [customerQueryKeys.cart, customerQueryKeys.ordersRoot];
    case REALTIME_EVENT_TYPE.orderStatusChanged: {
      const orderId = event.data.order_id;
      return [
        customerQueryKeys.ordersRoot,
        orderId ? customerQueryKeys.order(orderId) : customerQueryKeys.orderRoot,
      ];
    }
    case REALTIME_EVENT_TYPE.settingsUpdated:
      return [customerQueryKeys.pickupRules];
    default:
      return [];
  }
}
