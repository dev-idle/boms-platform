import { describe, expect, it } from "vitest";

import { reviewableProducts } from "./reviewable-products";

const croissant = "5d0c3f1e-9a4b-4c2d-8e7f-1a2b3c4d5e6f";
const cake = "7e1d4a2f-0b5c-4d3e-9f8a-2b3c4d5e6f7a";

describe("reviewable products", () => {
  it("lists each product picked up once, as the order lists them", () => {
    expect(
      reviewableProducts({
        status: "fulfilled",
        items: [
          { product_id: croissant, name: "Croissant" },
          { product_id: null, name: "Breakfast combo" },
          { product_id: cake, name: "Matcha cake" },
          { product_id: cake, name: "Matcha cake" },
        ],
      }),
    ).toEqual([
      { id: croissant, name: "Croissant" },
      { id: cake, name: "Matcha cake" },
    ]);
  });

  it("offers nothing before the order is picked up", () => {
    for (const status of ["confirmed", "ready", "cancelled", "no_show"] as const) {
      expect(reviewableProducts({ status, items: [{ product_id: croissant, name: "Croissant" }] })).toEqual([]);
    }
  });
});
