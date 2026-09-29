import { describe, expect, it } from "vitest";

import { kitchenTicketSchema } from "./index";

const ticket = {
  id: "00000000-0000-4000-8000-000000000001",
  order_id: "00000000-0000-4000-8000-000000000002",
  order_code: "CH-260928-001",
  order_status: "in_production",
  station: "kitchen",
  status: "in_progress",
  customer: { display_name: "Mai" },
  items: [{ name: "Matcha cake", quantity: 1 }],
  created_at: "2026-09-28T09:00:00+07:00",
};

describe("kitchen ticket", () => {
  it("reads where the whole order stands", () => {
    const result = kitchenTicketSchema.parse({
      ...ticket,
      order_tickets: [
        { station: "kitchen", status: "in_progress" },
        { station: "counter", status: "ready" },
      ],
    });
    expect(result.order_tickets).toHaveLength(2);
  });

  it("rejects a station it does not know", () => {
    expect(
      kitchenTicketSchema.safeParse({ ...ticket, order_tickets: [{ station: "bar", status: "ready" }] }).success,
    ).toBe(false);
  });
});
