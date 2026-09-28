import { describe, expect, it } from "vitest";

import { staffOrderSchema } from "./index";

const order = {
  id: "00000000-0000-4000-8000-000000000001",
  code: "CH-260928-001",
  status: "confirmed",
  subtotal_cents: 4500,
  discount_cents: 0,
  total_cents: 4500,
  items: [],
  customer: { user_id: "00000000-0000-4000-8000-000000000002", email: "mai@example.com" },
  created_at: "2026-09-28T09:00:00+07:00",
  updated_at: "2026-09-28T09:05:00+07:00",
};

// The counter's order history names the role behind every move.
describe("staff order timeline", () => {
  it("reads each move with the role that made it", () => {
    const result = staffOrderSchema.safeParse({
      ...order,
      timeline: [
        { status: "pending", actor_role: "customer", at: "2026-09-28T09:00:00+07:00" },
        { status: "confirmed", actor_role: "staff", at: "2026-09-28T09:05:00+07:00" },
      ],
    });
    expect(result.success).toBe(true);
  });

  it("rejects a move without a known role", () => {
    const withoutRole = staffOrderSchema.safeParse({
      ...order,
      timeline: [{ status: "pending", at: "2026-09-28T09:00:00+07:00" }],
    });
    const unknownRole = staffOrderSchema.safeParse({
      ...order,
      timeline: [{ status: "pending", actor_role: "system", at: "2026-09-28T09:00:00+07:00" }],
    });
    expect(withoutRole.success).toBe(false);
    expect(unknownRole.success).toBe(false);
  });
});
