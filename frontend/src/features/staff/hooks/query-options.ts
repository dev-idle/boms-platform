import type { QueryKey } from "@tanstack/react-query";

import { REALTIME_EVENT_TYPE, type RealtimeEvent } from "@/lib/realtime/events";
import type { StationTicketsListFilterInput } from "@/lib/schemas/ticket";

import type { StaffOrdersListFilterInput, StaffPickupsFilterInput } from "../schemas";

export const staffQueryKeys = {
  root: ["staff"] as const,
  ordersRoot: ["staff", "orders"] as const,
  orders: (filter: StaffOrdersListFilterInput) =>
    [...staffQueryKeys.ordersRoot, filter] as const,
  order: (id: string) => [...staffQueryKeys.ordersRoot, id] as const,
  pickups: (filter: StaffPickupsFilterInput) =>
    [...staffQueryKeys.ordersRoot, "pickups", filter] as const,
  ticketsRoot: ["staff", "tickets"] as const,
  tickets: (filter: StationTicketsListFilterInput) =>
    [...staffQueryKeys.ticketsRoot, filter] as const,
};

/** Everything pushed events can change in a staff tab, refetched after a gap. */
export const staffLiveQueryKeys: readonly QueryKey[] = [
  staffQueryKeys.ordersRoot,
  staffQueryKeys.ticketsRoot,
];

/**
 * Queries an event makes stale in a staff tab. The order queue is refetched
 * whole because a status change moves an order between its filters, and a paid
 * order arrives as one; a moved pickup also shows in the prep queue, a refund
 * only on the order; the prep
 * queue changes when an order is accepted, handed over or cancelled, and when
 * a ticket moves. A ticket also changes its order's detail.
 */
export function staffQueryKeysForEvent(event: RealtimeEvent): QueryKey[] {
  switch (event.type) {
    case REALTIME_EVENT_TYPE.orderStatusChanged:
    case REALTIME_EVENT_TYPE.orderRescheduled:
      return [staffQueryKeys.ordersRoot, staffQueryKeys.ticketsRoot];
    case REALTIME_EVENT_TYPE.orderRefunded:
      return [staffQueryKeys.ordersRoot];
    case REALTIME_EVENT_TYPE.ticketChanged: {
      const orderId = event.data.order_id;
      return [
        staffQueryKeys.ticketsRoot,
        orderId ? staffQueryKeys.order(orderId) : staffQueryKeys.ordersRoot,
      ];
    }
    default:
      return [];
  }
}
