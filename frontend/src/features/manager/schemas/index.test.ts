import { describe, expect, it } from "vitest";

import {
  categoryFormSchema,
  comboFormSchema,
  discountCodeFormSchema,
  productFormSchema,
} from "./index";

/** The shape a valid product form submits, so each case changes one field. */
const product = {
  category_id: "0f2f3a6e-5f4c-4f2e-9c1a-1d2b3c4d5e6f",
  name: "Almond croissant",
  slug: "almond-croissant",
  description: null,
  price_cents: 450,
  is_active: true,
  lead_time_minutes: 0,
  image_urls: [],
};

function priceIssue(value: unknown): string | undefined {
  const result = productFormSchema.safeParse({ ...product, price_cents: value });
  return result.success ? undefined : result.error.issues[0]?.message;
}

describe("catalog price", () => {
  it("is required, because the field starts empty rather than at zero", () => {
    expect(priceIssue(undefined)).toBe("Price is required");
  });

  it("names the bound it broke instead of reporting a union", () => {
    expect(priceIssue(-1)).toBe("Price must be zero or greater");
    expect(priceIssue(12.5)).toBe("Price must be zero or greater");
  });

  it("accepts free, which the catalog allows", () => {
    expect(priceIssue(0)).toBeUndefined();
  });

  it("calls anything that is not a number missing, not malformed", () => {
    expect(priceIssue("4.50")).toBe("Price is required");
    expect(priceIssue(Number.NaN)).toBe("Price is required");
  });

  it("guards the combo price the same way", () => {
    const result = comboFormSchema.safeParse({
      name: "Weekend box",
      slug: "weekend-box",
      image_url: "",
      starts_at: "2026-09-27T00:00:00Z",
      ends_at: "2026-10-04T00:00:00Z",
      is_active: true,
      items: [
        { product_id: "0f2f3a6e-5f4c-4f2e-9c1a-1d2b3c4d5e6f", quantity: 2 },
      ],
    });
    expect(result.success).toBe(false);
    expect(result.success ? [] : result.error.issues.map((i) => i.message)).toContain(
      "Price is required",
    );
  });

  it("guards the discount value with its own words", () => {
    const result = discountCodeFormSchema.safeParse({
      code: "SPRING",
      discount_type: "percent",
      min_order_cents: null,
      max_uses: null,
      max_discount_cents: null,
      starts_at: "2026-09-27T00:00:00Z",
      ends_at: "2026-10-04T00:00:00Z",
      is_active: true,
    });
    expect(result.success).toBe(false);
    expect(result.success ? [] : result.error.issues.map((i) => i.message)).toContain(
      "Value is required",
    );
  });
});

/**
 * The integer fields hand the form a number (or null when left empty), so the
 * schemas take numbers as they come instead of coercing whatever arrives.
 */
describe("catalog integer fields", () => {
  it("accept the number the sort-order field produces", () => {
    const result = categoryFormSchema.safeParse({
      name: "Breads",
      slug: "breads",
      sort_order: 0,
      is_active: true,
      station: "counter",
    });
    expect(result.success).toBe(true);
  });

  it("reject a string instead of reading a number out of it", () => {
    const result = categoryFormSchema.safeParse({
      name: "Breads",
      slug: "breads",
      sort_order: "3",
      is_active: true,
      station: "kitchen",
    });
    expect(result.success).toBe(false);
    expect(result.success ? [] : result.error.issues.map((i) => i.path)).toEqual([["sort_order"]]);
  });

  it("leave the optional discount limits empty as null", () => {
    const result = discountCodeFormSchema.safeParse({
      code: "SPRING",
      discount_type: "fixed_cents",
      value: 500,
      min_order_cents: null,
      max_uses: null,
      max_discount_cents: null,
      starts_at: "2026-09-27T00:00:00Z",
      ends_at: "2026-10-04T00:00:00Z",
      is_active: true,
    });
    expect(result.success).toBe(true);
  });
});

describe("catalog integer limits", () => {
  it("say a sort order past what the API stores is too large", () => {
    const result = categoryFormSchema.safeParse({
      name: "Breads",
      slug: "breads",
      sort_order: 2_147_483_648,
      is_active: true,
    });
    expect(result.success ? undefined : result.error.issues[0]?.message).toBe(
      "Sort order is too large",
    );
  });
});

describe("fulfillment fields", () => {
  it("name where a category is made", () => {
    const category = { name: "Breads", slug: "breads", sort_order: 0, is_active: true };
    expect(categoryFormSchema.safeParse({ ...category, station: "kitchen" }).success).toBe(true);
    expect(categoryFormSchema.safeParse({ ...category, station: "bar" }).success).toBe(false);
    expect(categoryFormSchema.safeParse(category).success).toBe(false);
  });

  it("bound the notice a product needs to a week", () => {
    expect(productFormSchema.safeParse({ ...product, lead_time_minutes: 10080 }).success).toBe(true);
    expect(productFormSchema.safeParse({ ...product, lead_time_minutes: 10081 }).success).toBe(false);
    expect(productFormSchema.safeParse({ ...product, lead_time_minutes: -1 }).success).toBe(false);
  });
});
