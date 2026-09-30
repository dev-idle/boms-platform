import { describe, expect, it } from "vitest";

import { ordersListFilterSchema, paymentStartSchema } from "./index";

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
