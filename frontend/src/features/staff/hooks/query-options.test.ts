import { describe, expect, it } from "vitest";

import { staffQueryKeys, staffQueryKeysForEvent } from "./query-options";

describe("staffQueryKeysForEvent", () => {
  it("refreshes the queue and order details for every order event", () => {
    for (const type of ["order.created", "order.status_changed"]) {
      expect(staffQueryKeysForEvent({ type, data: { order_id: "o-1" } })).toEqual([
        staffQueryKeys.ordersRoot,
      ]);
    }
    expect(staffQueryKeys.order("o-1").slice(0, 2)).toEqual([
      ...staffQueryKeys.ordersRoot,
    ]);
  });

  it("ignores events it does not know", () => {
    expect(staffQueryKeysForEvent({ type: "catalog.updated", data: {} })).toEqual([]);
  });
});
