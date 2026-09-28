import { describe, expect, it } from "vitest";

import { bakerQueryKeys, bakerQueryKeysForEvent } from "./query-options";

describe("bakerQueryKeysForEvent", () => {
  it("refreshes the board and order details for a status change", () => {
    expect(
      bakerQueryKeysForEvent({ type: "order.status_changed", data: { order_id: "o-1" } }),
    ).toEqual([bakerQueryKeys.productionRoot]);
    expect(bakerQueryKeys.productionOrder("o-1").slice(0, 2)).toEqual([
      ...bakerQueryKeys.productionRoot,
    ]);
  });

  it("ignores new orders, which are not production work yet", () => {
    expect(
      bakerQueryKeysForEvent({ type: "order.created", data: { order_id: "o-1" } }),
    ).toEqual([]);
  });
});
