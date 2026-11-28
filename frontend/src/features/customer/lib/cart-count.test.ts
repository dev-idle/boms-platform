import { describe, expect, it } from "vitest";

import { cartItemCount } from "./cart-count";

describe("cartItemCount", () => {
  it("adds the quantity of every line", () => {
    expect(cartItemCount([{ quantity: 1 }, { quantity: 4 }, { quantity: 2 }])).toBe(7);
  });

  it("counts nothing in an empty cart", () => {
    expect(cartItemCount([])).toBe(0);
  });
});
