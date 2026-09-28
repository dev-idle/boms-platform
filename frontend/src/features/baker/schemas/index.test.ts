import { describe, expect, it } from "vitest";

import { bakerOrderSummarySchema } from "./index";

const summary = {
  id: "00000000-0000-4000-8000-000000000001",
  status: "confirmed",
  total_cents: 4500,
  item_count: 2,
  pickup_at: "2026-09-29T09:00:00+07:00",
  created_at: "2026-09-28T09:00:00+07:00",
};

// The kitchen API sends a name for the order and no way to reach the customer.
describe("baker order customer", () => {
  it("reads an order that carries only the customer's display name", () => {
    const result = bakerOrderSummarySchema.safeParse({ ...summary, customer: { display_name: "Mai" } });
    expect(result.success).toBe(true);
  });

  it("reads an order whose customer gave no name", () => {
    const result = bakerOrderSummarySchema.safeParse({ ...summary, customer: {} });
    expect(result.success).toBe(true);
  });

  it("drops contact details if a response ever carried them", () => {
    const result = bakerOrderSummarySchema.parse({
      ...summary,
      customer: { display_name: "Mai", email: "mai@example.com" },
    });
    expect(result.customer).toEqual({ display_name: "Mai" });
  });
});
