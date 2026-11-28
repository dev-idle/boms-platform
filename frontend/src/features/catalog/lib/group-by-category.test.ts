import { describe, expect, it } from "vitest";

import type { CatalogProduct } from "@/lib/schemas/catalog";

import { groupByCategory } from "./group-by-category";

const CAKES = "0b9f7c0e-4c55-4a5e-9a3e-0f6b2a6f1a01";
const PASTRIES = "0b9f7c0e-4c55-4a5e-9a3e-0f6b2a6f1a02";

function product(id: string, categoryId: string, categoryName: string): CatalogProduct {
  return {
    id,
    category_id: categoryId,
    category_name: categoryName,
    category_slug: categoryName.toLowerCase(),
    name: `Product ${id}`,
    slug: `product-${id}`,
    price_cents: 500,
    image_urls: [],
    is_customizable: false,
    sold_out_today: false,
    options: [],
  };
}

describe("groupByCategory", () => {
  it("keeps the catalog's order of categories and of products within each", () => {
    const sections = groupByCategory([
      product("a", CAKES, "Cakes"),
      product("b", CAKES, "Cakes"),
      product("c", PASTRIES, "Pastries"),
    ]);
    expect(sections.map((section) => section.name)).toEqual(["Cakes", "Pastries"]);
    expect(sections[0].products.map((item) => item.id)).toEqual(["a", "b"]);
    expect(sections[1].id).toBe(PASTRIES);
  });

  it("gives no sections for no products", () => {
    expect(groupByCategory([])).toEqual([]);
  });
});
