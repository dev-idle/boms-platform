import { describe, expect, it } from "vitest";

import {
  PROMOTION_BODY_MAX_LENGTH,
  categoryFormSchema,
  comboFormSchema,
  discountCodeFormSchema,
  engagementReportSchema,
  incidentSummarySchema,
  managerIncidentSchema,
  managerPromotionSchema,
  productFormSchema,
  promotionFormSchema,
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
  is_customizable: false,
  options: [],
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

  it("checks each option a customizable product offers as the backend does", () => {
    const option = { group: "size", label: "20 cm", price_delta_cents: 1000, is_active: true };
    const custom = { ...product, is_customizable: true };
    expect(productFormSchema.safeParse({ ...custom, options: [option] }).success).toBe(true);
    for (const bad of [
      { ...option, label: "  " },
      { ...option, label: "a".repeat(61) },
      { ...option, label: "20\ncm" },
      { ...option, group: "shape" },
      { ...option, price_delta_cents: 100_001 },
    ]) {
      expect(productFormSchema.safeParse({ ...custom, options: [bad] }).success).toBe(false);
    }
  });
});

describe("promotionFormSchema", () => {
  it("trims what the manager wrote and asks for both parts", () => {
    expect(promotionFormSchema.parse({ subject: "  Matcha week ", body: " Ten percent off\n" })).toEqual({
      subject: "Matcha week",
      body: "Ten percent off",
    });
    expect(promotionFormSchema.safeParse({ subject: " ", body: "Hi" }).success).toBe(false);
    expect(promotionFormSchema.safeParse({ subject: "Hi", body: " \n " }).success).toBe(false);
  });

  it("counts characters as the API does, an emoji as one", () => {
    const longest = "🎂".repeat(PROMOTION_BODY_MAX_LENGTH);
    expect(promotionFormSchema.safeParse({ subject: "Hi", body: longest }).success).toBe(true);
    expect(promotionFormSchema.safeParse({ subject: "Hi", body: `${longest}a` }).success).toBe(false);
  });
});

describe("managerPromotionSchema", () => {
  it("reads a promotion still sending, before anyone is counted", () => {
    const sending = {
      id: "0b6c8f5e-3c1d-4a8e-9f0a-2d7e6c5b4a39",
      subject: "Matcha week",
      status: "sending",
      recipient_count: null,
      created_at: "2026-10-02T09:00:00+07:00",
      sender_name: "Hoa Pham",
    };
    expect(managerPromotionSchema.safeParse(sending).success).toBe(true);
    expect(managerPromotionSchema.safeParse({ ...sending, status: "draft" }).success).toBe(false);
  });
});

describe("managerIncidentSchema", () => {
  const reported = {
    id: "5f0c2b8e-7a4d-4c1e-9b3f-6d8e2a1c4b70",
    type: "wrong_items",
    source: "manual",
    note: "Two croissants short",
    created_at: "2026-10-02T10:15:00+07:00",
    order_code: "CH-261002-004",
    pickup_at: "2026-10-02T10:00:00+07:00",
    actor_name: "Lan Tran",
  };

  it("reads what staff reported and what the system recorded", () => {
    expect(managerIncidentSchema.safeParse(reported).success).toBe(true);
    expect(
      managerIncidentSchema.safeParse({
        ...reported,
        type: "payment_anomaly",
        source: "auto",
        note: null,
        pickup_at: null,
        actor_name: null,
      }).success,
    ).toBe(true);
  });

  it("refuses a type or source the API does not send", () => {
    expect(managerIncidentSchema.safeParse({ ...reported, type: "late" }).success).toBe(false);
    expect(managerIncidentSchema.safeParse({ ...reported, source: "staff" }).success).toBe(false);
  });
});

describe("incidentSummarySchema", () => {
  it("counts each type a week holds; a type with none is left out", () => {
    expect(incidentSummarySchema.safeParse({ types: [{ type: "no_show", count: 2 }] }).success).toBe(true);
    expect(incidentSummarySchema.safeParse({ types: [{ type: "no_show", count: 0 }] }).success).toBe(false);
  });
});

describe("engagementReportSchema", () => {
  const report = {
    customers: 4,
    used_none: 1,
    used_one: 1,
    used_two_or_more: 2,
    features: [
      { feature: "reviews", customers: 1 },
      { feature: "promotions", customers: 0 },
    ],
  };

  it("reads the customers, their spread and each feature", () => {
    expect(engagementReportSchema.safeParse(report).success).toBe(true);
    expect(engagementReportSchema.safeParse({ ...report, customers: 0, used_none: 0, used_one: 0, used_two_or_more: 0 }).success).toBe(true);
  });

  it("refuses a feature the API does not report", () => {
    expect(
      engagementReportSchema.safeParse({ ...report, features: [{ feature: "chatbot", customers: 1 }] }).success,
    ).toBe(false);
  });
});
