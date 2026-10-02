import { describe, expect, it } from "vitest";

import {
  cartCustomizationInputSchema,
  orderSchema,
  ordersListFilterSchema,
  paymentStartSchema,
  reviewInputSchema,
  savedProductSchema,
} from "./index";

describe("ordersListFilterSchema", () => {
  it("accepts an open or a closed range of days", () => {
    expect(ordersListFilterSchema.safeParse({}).success).toBe(true);
    expect(ordersListFilterSchema.safeParse({ from: "2026-09-01" }).success).toBe(true);
    expect(
      ordersListFilterSchema.safeParse({ from: "2026-09-28", to: "2026-09-28", status: "ready" }).success,
    ).toBe(true);
  });

  it("rejects a range that ends before it starts", () => {
    const result = ordersListFilterSchema.safeParse({ from: "2026-09-28", to: "2026-09-27" });
    expect(result.success).toBe(false);
    expect(result.error?.issues[0]?.path).toEqual(["to"]);
  });

  it("rejects days that are not YYYY-MM-DD", () => {
    expect(ordersListFilterSchema.safeParse({ from: "28/09/2026" }).success).toBe(false);
  });
});

// The browser is sent to this page, so only PayPal's own may be it.
describe("paymentStartSchema", () => {
  it("accepts a PayPal page", () => {
    for (const url of [
      "https://www.sandbox.paypal.com/checkoutnow?token=5O190127TN364715T",
      "https://www.paypal.com/checkoutnow?token=5O190127TN364715T",
    ]) {
      expect(paymentStartSchema.safeParse({ approve_url: url }).success).toBe(true);
    }
  });

  it("refuses any other page", () => {
    for (const url of [
      "http://www.paypal.com/checkoutnow",
      "https://paypal.com.evil.example/checkoutnow",
      "https://evilpaypal.com/checkoutnow",
      "javascript:alert(1)",
    ]) {
      expect(paymentStartSchema.safeParse({ approve_url: url }).success, url).toBe(false);
    }
  });
});

// A custom cake is checked as the backend checks it before it goes in the cart.
describe("cartCustomizationInputSchema", () => {
  const cake = {
    option_ids: ["0f2f3a6e-5f4c-4f2e-9c1a-1d2b3c4d5e6f"],
    message: " Happy birthday, Mai ",
    reference_image_url: "",
    reference_rights_confirmed: false,
  };

  it("keeps a plain message, trimmed", () => {
    expect(cartCustomizationInputSchema.parse(cake).message).toBe("Happy birthday, Mai");
  });

  it("asks for an option in every group offered", () => {
    expect(cartCustomizationInputSchema.safeParse({ ...cake, option_ids: [""] }).success).toBe(false);
  });

  it("refuses a long or formatted message", () => {
    for (const message of ["a".repeat(61), "Happy\nbirthday"]) {
      expect(cartCustomizationInputSchema.safeParse({ ...cake, message }).success).toBe(false);
    }
  });

  it("takes a photo only with the customer's word that they may share it", () => {
    const photo = { ...cake, reference_image_url: "https://res.cloudinary.com/demo/image/upload/v1/boms/references/u/cake.jpg" };
    expect(cartCustomizationInputSchema.safeParse(photo).success).toBe(false);
    expect(cartCustomizationInputSchema.safeParse({ ...photo, reference_rights_confirmed: true }).success).toBe(true);
  });
});

// The customer gives this code at the counter, so it is exactly what the counter types.
describe("orderSchema pickup code", () => {
  const order = {
    id: "00000000-0000-4000-8000-000000000001",
    code: "CH-261001-001",
    status: "confirmed",
    order_type: "pre_order",
    subtotal_cents: 4500,
    discount_cents: 0,
    total_cents: 4500,
    items: [],
    timeline: [],
    tickets: [],
    payment: null,
    fulfillment: null,
    payment_due_at: null,
    can_message: true,
    created_at: "2026-10-01T09:00:00+07:00",
    updated_at: "2026-10-01T09:05:00+07:00",
  };

  it("reads four digits, or none before the bakery accepts the order", () => {
    expect(orderSchema.parse({ ...order, pickup_code: "0427" }).pickup_code).toBe("0427");
    expect(orderSchema.parse({ ...order, pickup_code: null }).pickup_code).toBeNull();
  });

  it("rejects anything else", () => {
    for (const pickup_code of [undefined, "427", "04a7"]) {
      expect(orderSchema.safeParse({ ...order, pickup_code }).success).toBe(false);
    }
  });
});

describe("savedProductSchema", () => {
  const saved = {
    list: "wishlist",
    saved_at: "2026-10-02T09:00:00+07:00",
    product: {
      id: "0b6c8f5e-3c1d-4a8e-9f0a-2d7e6c5b4a39",
      category_id: "6f1e2d3c-4b5a-4987-8a6b-5c4d3e2f1a0b",
      category_name: "Pastries",
      category_slug: "pastries",
      name: "Croissant",
      slug: "croissant",
      price_cents: 380,
      image_urls: [],
      is_customizable: false,
      sold_out_today: false,
    },
  };

  it("reads a product on a list as the catalog shows it", () => {
    expect(savedProductSchema.parse(saved).product.name).toBe("Croissant");
  });

  it("knows only the favorites and the wishlist", () => {
    expect(savedProductSchema.safeParse({ ...saved, list: "basket" }).success).toBe(false);
  });
});

describe("reviewInputSchema", () => {
  const productId = "0b6c8f5e-3c1d-4a8e-9f0a-2d7e6c5b4a39";

  it("asks for a rating before anything is posted", () => {
    expect(reviewInputSchema.safeParse({ product_id: productId, rating: 0, comment: "" }).success).toBe(false);
    expect(reviewInputSchema.safeParse({ product_id: productId, rating: 6, comment: "" }).success).toBe(false);
  });

  it("keeps the comment trimmed, and a rating alone", () => {
    expect(reviewInputSchema.parse({ product_id: productId, rating: 5, comment: "  Flaky\nand buttery " }).comment).toBe(
      "Flaky\nand buttery",
    );
    expect(reviewInputSchema.parse({ product_id: productId, rating: 4, comment: "  " }).comment).toBe("");
  });

  it("counts the comment's characters as the API does, an emoji as one", () => {
    const longest = "🎂".repeat(1000);
    expect(reviewInputSchema.safeParse({ product_id: productId, rating: 5, comment: longest }).success).toBe(true);
    expect(reviewInputSchema.safeParse({ product_id: productId, rating: 5, comment: `${longest}a` }).success).toBe(false);
  });
});
