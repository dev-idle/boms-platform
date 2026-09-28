import { describe, expect, it } from "vitest";

import { ordersListFilterSchema } from "./index";

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
