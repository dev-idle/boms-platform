import { describe, expect, it } from "vitest";

import { customerQueryKeys, customerQueryKeysForEvent } from "./query-options";

describe("customerQueryKeysForEvent", () => {
  it("refreshes the cart and order list for a new order", () => {
    expect(
      customerQueryKeysForEvent({ type: "order.created", data: { order_id: "o-1" } }),
    ).toEqual([customerQueryKeys.cart, customerQueryKeys.ordersRoot]);
  });

  it("refreshes the order list and that order for a status change", () => {
    expect(
      customerQueryKeysForEvent({
        type: "order.status_changed",
        data: { order_id: "o-1", status: "ready" },
      }),
    ).toEqual([customerQueryKeys.ordersRoot, customerQueryKeys.order("o-1")]);
  });

  it("refreshes every open order when the event names none", () => {
    expect(
      customerQueryKeysForEvent({ type: "order.status_changed", data: {} }),
    ).toEqual([customerQueryKeys.ordersRoot, customerQueryKeys.orderRoot]);
  });

  it("ignores events it does not know", () => {
    expect(customerQueryKeysForEvent({ type: "settings.updated", data: {} })).toEqual([]);
  });
});
