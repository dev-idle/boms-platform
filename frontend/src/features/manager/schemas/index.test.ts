import { describe, expect, it } from "vitest";

import { comboFormSchema, discountCodeFormSchema, productFormSchema } from "./index";

/** The shape a valid product form submits, so each case changes one field. */
const product = {
  category_id: "0f2f3a6e-5f4c-4f2e-9c1a-1d2b3c4d5e6f",
  name: "Almond croissant",
  slug: "almond-croissant",
  description: null,
  price_cents: 450,
  is_active: true,
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
