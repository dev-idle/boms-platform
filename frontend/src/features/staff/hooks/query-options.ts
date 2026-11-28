import type { QueryKey } from "@tanstack/react-query";

import { pickupQueryKeys } from "@/features/customer";
import { REALTIME_EVENT_TYPE, type RealtimeEvent } from "@/lib/realtime/events";

import type {
  StaffInboxFilterInput,
  StaffOrderItemInput,
  StaffOrdersListFilterInput,
  StaffPickupsFilterInput,
  StaffProductsFilterInput,
  StaffThreadRef,
} from "../schemas";

export const staffQueryKeys = {
  root: ["staff"] as const,
  ordersRoot: ["staff", "orders"] as const,
  orders: (filter: StaffOrdersListFilterInput) =>
    [...staffQueryKeys.ordersRoot, filter] as const,
  order: (id: string) => [...staffQueryKeys.ordersRoot, id] as const,
  pickupsRoot: ["staff", "orders", "pickups"] as const,
  pickups: (filter: StaffPickupsFilterInput) => [...staffQueryKeys.pickupsRoot, filter] as const,
  productsRoot: ["staff", "products"] as const,
  products: (filter: StaffProductsFilterInput) => [...staffQueryKeys.productsRoot, filter] as const,
  quote: (items: StaffOrderItemInput[]) => ["staff", "quote", items] as const,
  customer: (email: string) => ["staff", "customer", email] as const,
  conversationsRoot: ["staff", "conversations"] as const,
  conversations: (filter: StaffInboxFilterInput) => [...staffQueryKeys.conversationsRoot, filter] as const,
  conversationCounts: ["staff", "conversations", "counts"] as const,
  messagesRoot: ["staff", "messages"] as const,
  messages: (thread: StaffThreadRef) => [...staffQueryKeys.messagesRoot, thread.kind, thread.id] as const,
};

/** Everything pushed events can change in a staff tab, refetched after a gap. */
export const staffLiveQueryKeys: readonly QueryKey[] = [
  staffQueryKeys.ordersRoot,
  staffQueryKeys.productsRoot,
  staffQueryKeys.conversationsRoot,
  staffQueryKeys.messagesRoot,
  ...pickupQueryKeys,
];

/**
 * Queries an event makes stale in a staff tab. The order queue, the pickup
 * schedule among it, is refetched whole because a status change moves an
 * order between its filters, and a paid order arrives as one; a refund shows
 * only on the order. A ticket moving changes the schedule, where the counter
 * gets its part ready, and its order's detail. An order taken at the counter
 * arrives confirmed, as one paid does. A product running out, or coming back,
 * changes the availability list. A message arriving, or a conversation read or
 * resolved, changes the inbox, its counts and the thread — an order's, or a
 * customer's general conversation.
 */
export function staffQueryKeysForEvent(event: RealtimeEvent): QueryKey[] {
  switch (event.type) {
    case REALTIME_EVENT_TYPE.orderCreated:
    case REALTIME_EVENT_TYPE.orderStatusChanged:
    case REALTIME_EVENT_TYPE.orderRescheduled:
      return [staffQueryKeys.ordersRoot];
    case REALTIME_EVENT_TYPE.productSoldOutChanged:
      return [staffQueryKeys.productsRoot];
    case REALTIME_EVENT_TYPE.messageCreated:
    case REALTIME_EVENT_TYPE.conversationChanged: {
      const { order_id: orderId, customer_id: customerId } = event.data;
      if (orderId) {
        return [staffQueryKeys.conversationsRoot, staffQueryKeys.messages({ kind: "order", id: orderId })];
      }
      if (customerId) {
        return [staffQueryKeys.conversationsRoot, staffQueryKeys.messages({ kind: "customer", id: customerId })];
      }
      return [staffQueryKeys.conversationsRoot, staffQueryKeys.messagesRoot];
    }
    case REALTIME_EVENT_TYPE.orderRefunded:
      return [staffQueryKeys.ordersRoot];
    // The picker in a new order or a moved pickup: a slot filling, or the rules changing.
    case REALTIME_EVENT_TYPE.slotsChanged:
    case REALTIME_EVENT_TYPE.settingsUpdated:
      return [...pickupQueryKeys];
    case REALTIME_EVENT_TYPE.ticketChanged: {
      const orderId = event.data.order_id;
      return orderId
        ? [staffQueryKeys.pickupsRoot, staffQueryKeys.order(orderId)]
        : [staffQueryKeys.ordersRoot];
    }
    default:
      return [];
  }
}
