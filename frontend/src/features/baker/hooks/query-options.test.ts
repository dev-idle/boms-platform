import { describe, expect, it } from "vitest";

import { bakerQueryKeys, bakerQueryKeysForEvent } from "./query-options";

describe("bakerQueryKeysForEvent", () => {
  it("refreshes the queue and its tickets when a ticket, its order or its pickup moves", () => {
    for (const type of ["ticket.changed", "order.status_changed", "order.rescheduled"]) {
      expect(bakerQueryKeysForEvent({ type, data: { order_id: "o-1" } })).toEqual([
        bakerQueryKeys.ticketsRoot,
      ]);
    }
    expect(bakerQueryKeys.ticket("t-1").slice(0, 2)).toEqual([...bakerQueryKeys.ticketsRoot]);
  });

  it("ignores new orders, which are not production work yet", () => {
    expect(
      bakerQueryKeysForEvent({ type: "order.created", data: { order_id: "o-1" } }),
    ).toEqual([]);
  });
});
