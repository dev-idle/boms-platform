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

  it("keeps a day's pickups under the orders, so every order event refreshes them", () => {
    expect(staffQueryKeys.pickups({ date: "2026-10-01" }).slice(0, 2)).toEqual([...staffQueryKeys.ordersRoot]);
  });

  it("refreshes both queues when the counter takes an order, and the availability list when a product runs out", () => {
    expect(staffQueryKeysForEvent({ type: "order.created", data: { order_id: "o-1" } })).toEqual([
      staffQueryKeys.ordersRoot,
      staffQueryKeys.ticketsRoot,
    ]);
    expect(staffQueryKeysForEvent({ type: "product.sold_out_changed", data: { product_id: "p-1" } })).toEqual([
      staffQueryKeys.productsRoot,
    ]);
  });

  it("refreshes the inbox and the order's thread when a message arrives or a conversation changes", () => {
    for (const type of ["message.created", "conversation.changed"]) {
      expect(staffQueryKeysForEvent({ type, data: { order_id: "o-1" } })).toEqual([
        staffQueryKeys.conversationsRoot,
        staffQueryKeys.messages("o-1"),
      ]);
    }
    expect(staffQueryKeysForEvent({ type: "conversation.changed", data: {} })).toEqual([
      staffQueryKeys.conversationsRoot,
      staffQueryKeys.messagesRoot,
    ]);
  });

  it("ignores events it does not know", () => {
    expect(staffQueryKeysForEvent({ type: "catalog.updated", data: {} })).toEqual([]);
  });
});
