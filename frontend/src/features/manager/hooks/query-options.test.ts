import { describe, expect, it } from "vitest";

import { managerQueryKeys, managerQueryKeysForEvent } from "./query-options";

describe("managerQueryKeysForEvent", () => {
  it("refreshes the reviews and their summary when one is written, published or hidden", () => {
    expect(managerQueryKeysForEvent({ type: "review.changed", data: { product_id: "p-1" } })).toEqual([
      managerQueryKeys.reviewsRoot,
    ]);
  });

  it("holds the summary under the reviews it adds up", () => {
    expect(managerQueryKeys.reviewSummary.slice(0, managerQueryKeys.reviewsRoot.length)).toEqual([
      ...managerQueryKeys.reviewsRoot,
    ]);
  });

  it("refreshes the promotions when one is sent or done sending", () => {
    for (const type of ["promotion.created", "promotion.sent"]) {
      expect(managerQueryKeysForEvent({ type, data: { promotion_id: "p-1" } })).toEqual([managerQueryKeys.promotionsRoot]);
    }
  });

  it("ignores events it does not know", () => {
    expect(managerQueryKeysForEvent({ type: "order.created", data: { order_id: "o-1" } })).toEqual([]);
  });
});
