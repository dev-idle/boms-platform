import { describe, expect, it } from "vitest";

import { staffQueryKeys, staffQueryKeysForEvent } from "./query-options";

describe("staffQueryKeysForEvent", () => {
  it("refreshes both queues and every order detail when an order moves or is paid", () => {
    expect(
      staffQueryKeysForEvent({ type: "order.status_changed", data: { order_id: "o-1" } }),
    ).toEqual([staffQueryKeys.ordersRoot, staffQueryKeys.ticketsRoot]);
    expect(staffQueryKeys.order("o-1").slice(0, 2)).toEqual([
      ...staffQueryKeys.ordersRoot,
    ]);
  });

  it("refreshes both queues when a pickup moves, and only the orders when a refund arrives", () => {
    expect(
      staffQueryKeysForEvent({ type: "order.rescheduled", data: { order_id: "o-1" } }),
    ).toEqual([staffQueryKeys.ordersRoot, staffQueryKeys.ticketsRoot]);
    expect(
      staffQueryKeysForEvent({ type: "order.refunded", data: { order_id: "o-1" } }),
    ).toEqual([staffQueryKeys.ordersRoot]);
  });

  it("refreshes the prep queue and the ticket's order when a ticket moves", () => {
    expect(
      staffQueryKeysForEvent({ type: "ticket.changed", data: { order_id: "o-1", ticket_id: "t-1" } }),
    ).toEqual([staffQueryKeys.ticketsRoot, staffQueryKeys.order("o-1")]);
  });

  it("refreshes every order detail when a ticket event names no order", () => {
    expect(staffQueryKeysForEvent({ type: "ticket.changed", data: {} })).toEqual([
      staffQueryKeys.ticketsRoot,
      staffQueryKeys.ordersRoot,
    ]);
  });

  it("ignores events it does not know", () => {
    expect(staffQueryKeysForEvent({ type: "catalog.updated", data: {} })).toEqual([]);
  });
});
