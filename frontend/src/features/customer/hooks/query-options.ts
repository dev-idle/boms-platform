import type { QueryKey } from "@tanstack/react-query";

import { REALTIME_EVENT_TYPE, type RealtimeEvent } from "@/lib/realtime/events";

import type { OrdersListFilterInput } from "../schemas";

export const customerQueryKeys = {
  cart: ["customer", "cart"] as const,
  pickupRules: ["customer", "pickup-rules"] as const,
  pickupSlotsRoot: ["customer", "pickup-slots"] as const,
  pickupSlots: (date: string) => [...customerQueryKeys.pickupSlotsRoot, date] as const,
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
  customerQueryKeys.pickupSlotsRoot,
  customerQueryKeys.ordersRoot,
  customerQueryKeys.orderRoot,
];

/**
 * Queries a pushed event makes stale in a customer tab. Checkout empties the
 * cart, so a new order placed in another tab refreshes the cart here too; a
 * settings change refreshes the pickup window the cart offers, and an order
 * taking or freeing a slot refreshes that day's slots. An order moving, its
 * pickup moving or its refund arriving changes the order list and its detail;
 * a ticket moving changes only its order's detail, where the preparation shows.
 * A product running out today, or coming back, changes the days the cart can
 * be collected on.
 */
export function customerQueryKeysForEvent(event: RealtimeEvent): QueryKey[] {
  switch (event.type) {
    case REALTIME_EVENT_TYPE.orderCreated:
      return [customerQueryKeys.cart, customerQueryKeys.ordersRoot];
    case REALTIME_EVENT_TYPE.orderStatusChanged:
    case REALTIME_EVENT_TYPE.orderRescheduled:
    case REALTIME_EVENT_TYPE.orderRefunded: {
      const orderId = event.data.order_id;
      return [
        customerQueryKeys.ordersRoot,
        orderId ? customerQueryKeys.order(orderId) : customerQueryKeys.orderRoot,
      ];
    }
    case REALTIME_EVENT_TYPE.ticketChanged: {
      const orderId = event.data.order_id;
      return [orderId ? customerQueryKeys.order(orderId) : customerQueryKeys.orderRoot];
    }
    case REALTIME_EVENT_TYPE.settingsUpdated:
      return [customerQueryKeys.pickupRules, customerQueryKeys.pickupSlotsRoot];
    case REALTIME_EVENT_TYPE.productSoldOutChanged:
      return [customerQueryKeys.cart];
    case REALTIME_EVENT_TYPE.slotsChanged: {
      const date = event.data.date;
      return [date ? customerQueryKeys.pickupSlots(date) : customerQueryKeys.pickupSlotsRoot];
    }
    default:
      return [];
  }
}
