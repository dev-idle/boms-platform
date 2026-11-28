import { describe, expect, it } from "vitest";

import type { OrderSummary } from "../schemas";
import { nextPickup, unreadOrderMessages } from "./orders-glance";

/** 10:00 on Nov 26, 2026 in the bakery (UTC+7). */
const NOW = new Date("2026-11-26T03:00:00Z");

function order(overrides: Partial<OrderSummary>): OrderSummary {
  return {
    id: "6f0f4f8e-2d1c-4f3a-9b1e-1a2b3c4d5e6f",
    code: "CH-261124-007",
    status: "confirmed",
    total_cents: 3900,
    item_count: 2,
    unread_messages: 0,
    pickup_at: "2026-11-28T07:00:00Z",
    created_at: "2026-11-24T01:15:00Z",
    ...overrides,
  };
}

describe("nextPickup", () => {
  it("picks the earliest pickup among orders still to collect", () => {
    const later = order({ code: "CH-261124-007", pickup_at: "2026-11-28T07:00:00Z" });
    const sooner = order({ code: "CH-261125-002", status: "ready", pickup_at: "2026-11-27T02:00:00Z" });
    expect(nextPickup([later, sooner], NOW)).toBe(sooner);
  });

  it("keeps a pickup earlier today, which may still wait at the counter", () => {
    const earlierToday = order({ status: "ready", pickup_at: "2026-11-26T01:00:00Z" });
    expect(nextPickup([earlierToday], NOW)).toBe(earlierToday);
  });

  it("drops a pickup from a past day", () => {
    expect(nextPickup([order({ pickup_at: "2026-10-02T01:00:00Z" })], NOW)).toBeNull();
  });

  it("ignores orders not taken yet, handed over or ended", () => {
    const orders = (["awaiting_payment", "fulfilled", "cancelled", "expired", "no_show"] as const).map((status) =>
      order({ status, pickup_at: "2026-11-27T02:00:00Z" }),
    );
    expect(nextPickup(orders, NOW)).toBeNull();
  });

  it("ignores an order without a pickup time", () => {
    expect(nextPickup([order({ pickup_at: null })], NOW)).toBeNull();
  });
});

describe("unreadOrderMessages", () => {
  it("adds the unread replies of every order", () => {
    expect(unreadOrderMessages([order({ unread_messages: 1 }), order({ unread_messages: 2 })])).toBe(3);
  });
});
