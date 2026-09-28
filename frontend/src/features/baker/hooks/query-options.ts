import type { QueryKey } from "@tanstack/react-query";

import { REALTIME_EVENT_TYPE, type RealtimeEvent } from "@/lib/realtime/events";

import type { BakerOrdersListFilterInput } from "../schemas";

export const bakerQueryKeys = {
  productionRoot: ["baker", "production"] as const,
  production: (filter: BakerOrdersListFilterInput) =>
    [...bakerQueryKeys.productionRoot, filter] as const,
  productionOrder: (id: string) => ["baker", "production", id] as const,
};

/** Everything pushed events can change in a baker tab, refetched after a gap. */
export const bakerLiveQueryKeys: readonly QueryKey[] = [bakerQueryKeys.productionRoot];

/**
 * Queries an order event makes stale in a baker tab: the production board and
 * the order's detail, both under the production root. The API only tells the
 * kitchen about orders entering, leaving or moving within its statuses.
 */
export function bakerQueryKeysForEvent(event: RealtimeEvent): QueryKey[] {
  switch (event.type) {
    case REALTIME_EVENT_TYPE.orderStatusChanged:
      return [bakerQueryKeys.productionRoot];
    default:
      return [];
  }
}
