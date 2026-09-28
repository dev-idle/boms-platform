import type { QueryKey } from "@tanstack/react-query";

import { REALTIME_EVENT_TYPE, type RealtimeEvent } from "@/lib/realtime/events";

import type { StaffOrdersListFilterInput } from "../schemas";

export const staffQueryKeys = {
  root: ["staff"] as const,
  ordersRoot: ["staff", "orders"] as const,
  orders: (filter: StaffOrdersListFilterInput) =>
    [...staffQueryKeys.ordersRoot, filter] as const,
  order: (id: string) => [...staffQueryKeys.ordersRoot, id] as const,
};

/** Everything pushed events can change in a staff tab, refetched after a gap. */
export const staffLiveQueryKeys: readonly QueryKey[] = [staffQueryKeys.ordersRoot];

/**
 * Queries an order event makes stale in a staff tab: the queue and the order's
 * detail, both under the orders root. The queue is refetched whole because a
 * status change moves an order between its filters.
 */
export function staffQueryKeysForEvent(event: RealtimeEvent): QueryKey[] {
  switch (event.type) {
    case REALTIME_EVENT_TYPE.orderCreated:
    case REALTIME_EVENT_TYPE.orderStatusChanged:
      return [staffQueryKeys.ordersRoot];
    default:
      return [];
  }
}
