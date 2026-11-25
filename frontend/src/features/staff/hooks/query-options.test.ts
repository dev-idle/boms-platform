import { describe, expect, it } from "vitest";

import { staffQueryKeys, staffQueryKeysForEvent } from "./query-options";

describe("staffQueryKeysForEvent", () => {
  it("refreshes the orders, the schedule and every order detail when an order moves or is paid", () => {
    expect(
      staffQueryKeysForEvent({ type: "order.status_changed", data: { order_id: "o-1" } }),
    ).toEqual([staffQueryKeys.ordersRoot]);
    expect(staffQueryKeys.order("o-1").slice(0, 2)).toEqual([
      ...staffQueryKeys.ordersRoot,
    ]);
  });

  it("refreshes the orders when a pickup moves or a refund arrives", () => {
    expect(
      staffQueryKeysForEvent({ type: "order.rescheduled", data: { order_id: "o-1" } }),
    ).toEqual([staffQueryKeys.ordersRoot]);
    expect(
      staffQueryKeysForEvent({ type: "order.refunded", data: { order_id: "o-1" } }),
    ).toEqual([staffQueryKeys.ordersRoot]);
  });

  it("refreshes the schedule and the ticket's order when a ticket moves", () => {
    expect(
      staffQueryKeysForEvent({ type: "ticket.changed", data: { order_id: "o-1", ticket_id: "t-1" } }),
    ).toEqual([staffQueryKeys.pickupsRoot, staffQueryKeys.order("o-1")]);
  });

  it("refreshes every order when a ticket event names no order", () => {
    expect(staffQueryKeysForEvent({ type: "ticket.changed", data: {} })).toEqual([staffQueryKeys.ordersRoot]);
  });

  it("keeps a day's pickups under the orders, so every order event refreshes them", () => {
    expect(staffQueryKeys.pickups({ date: "2026-10-01" }).slice(0, 3)).toEqual([...staffQueryKeys.pickupsRoot]);
    expect(staffQueryKeys.pickupsRoot.slice(0, 2)).toEqual([...staffQueryKeys.ordersRoot]);
  });

  it("refreshes the orders when the counter takes one, and the availability list when a product runs out", () => {
    expect(staffQueryKeysForEvent({ type: "order.created", data: { order_id: "o-1" } })).toEqual([
      staffQueryKeys.ordersRoot,
    ]);
    expect(staffQueryKeysForEvent({ type: "product.sold_out_changed", data: { product_id: "p-1" } })).toEqual([
      staffQueryKeys.productsRoot,
    ]);
  });

  it("refreshes the inbox and the thread when a message arrives or a conversation changes", () => {
    for (const type of ["message.created", "conversation.changed"]) {
      expect(staffQueryKeysForEvent({ type, data: { order_id: "o-1" } })).toEqual([
        staffQueryKeys.conversationsRoot,
        staffQueryKeys.messages({ kind: "order", id: "o-1" }),
      ]);
      expect(staffQueryKeysForEvent({ type, data: { customer_id: "c-1" } })).toEqual([
        staffQueryKeys.conversationsRoot,
        staffQueryKeys.messages({ kind: "customer", id: "c-1" }),
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
