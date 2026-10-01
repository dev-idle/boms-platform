import type { QueryKey } from "@tanstack/react-query";

import { REALTIME_EVENT_TYPE, type RealtimeEvent } from "@/lib/realtime/events";
import type { StationTicketsListFilterInput } from "@/lib/schemas/ticket";

export const bakerQueryKeys = {
  ticketsRoot: ["baker", "tickets"] as const,
  tickets: (filter: StationTicketsListFilterInput) =>
    [...bakerQueryKeys.ticketsRoot, filter] as const,
  ticket: (id: string) => [...bakerQueryKeys.ticketsRoot, id] as const,
};

/** Everything pushed events can change in a baker tab, refetched after a gap. */
export const bakerLiveQueryKeys: readonly QueryKey[] = [bakerQueryKeys.ticketsRoot];

/**
 * Queries an event makes stale in a baker tab: the kitchen queue and its
 * tickets, all under the tickets root. A ticket moves when a station works on
 * it; the queue also changes when the counter accepts, hands over or cancels an
 * order, which reaches the kitchen as the order's status change, when the
 * counter takes an order, which starts accepted, and when a customer moves the
 * pickup of an accepted order.
 */
export function bakerQueryKeysForEvent(event: RealtimeEvent): QueryKey[] {
  switch (event.type) {
    case REALTIME_EVENT_TYPE.orderCreated:
    case REALTIME_EVENT_TYPE.orderStatusChanged:
    case REALTIME_EVENT_TYPE.orderRescheduled:
    case REALTIME_EVENT_TYPE.ticketChanged:
      return [bakerQueryKeys.ticketsRoot];
    default:
      return [];
  }
}
