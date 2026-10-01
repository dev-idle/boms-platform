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

  it("refreshes the order list and that order when its pickup moves or its refund arrives", () => {
    for (const type of ["order.rescheduled", "order.refunded"]) {
      expect(customerQueryKeysForEvent({ type, data: { order_id: "o-1" } })).toEqual([
        customerQueryKeys.ordersRoot,
        customerQueryKeys.order("o-1"),
      ]);
    }
  });

  it("refreshes every open order when the event names none", () => {
    expect(
      customerQueryKeysForEvent({ type: "order.status_changed", data: {} }),
    ).toEqual([customerQueryKeys.ordersRoot, customerQueryKeys.orderRoot]);
  });

  it("refreshes only the order whose ticket moved", () => {
    expect(
      customerQueryKeysForEvent({ type: "ticket.changed", data: { order_id: "o-1", ticket_id: "t-1" } }),
    ).toEqual([customerQueryKeys.order("o-1")]);
    expect(customerQueryKeysForEvent({ type: "ticket.changed", data: {} })).toEqual([
      customerQueryKeys.orderRoot,
    ]);
  });

  it("ignores events it does not know", () => {
    expect(customerQueryKeysForEvent({ type: "catalog.updated", data: {} })).toEqual([]);
  });
});

describe("customerQueryKeysForEvent for settings", () => {
  it("refreshes the pickup window when an admin changes it", () => {
    expect(customerQueryKeysForEvent({ type: "settings.updated", data: {} })).toEqual([
      customerQueryKeys.pickupRules,
      customerQueryKeys.pickupSlotsRoot,
    ]);
  });

  it("refreshes the slots of the day an order took or freed", () => {
    expect(customerQueryKeysForEvent({ type: "slots.changed", data: { date: "2026-07-10" } })).toEqual([
      customerQueryKeys.pickupSlots("2026-07-10"),
    ]);
  });
});
