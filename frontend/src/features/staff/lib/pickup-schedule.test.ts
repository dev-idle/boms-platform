import { describe, expect, it } from "vitest";

import type { StaffPickup } from "../schemas";
import { groupPickupsBySlot, isLatePickup } from "./pickup-schedule";

const pickup = (id: number, pickupAt: string, status: StaffPickup["status"] = "confirmed"): StaffPickup => ({
  id: `00000000-0000-4000-8000-00000000000${id}`,
  code: `CH-261001-00${id}`,
  status,
  channel: "online",
  total_cents: 4500,
  item_count: 2,
  customer: { user_id: "00000000-0000-4000-8000-000000000009", email: "mai@example.com" },
  pickup_at: pickupAt,
  created_at: "2026-10-01T08:00:00+07:00",
});

describe("groupPickupsBySlot", () => {
  it("puts pickups at the same time in one slot, in order", () => {
    const slots = groupPickupsBySlot([
      pickup(1, "2026-10-01T09:00:00+07:00"),
      pickup(2, "2026-10-01T02:00:00Z"),
      pickup(3, "2026-10-01T09:30:00+07:00"),
    ]);
    expect(slots.map((slot) => slot.pickups.map((p) => p.code))).toEqual([
      ["CH-261001-001", "CH-261001-002"],
      ["CH-261001-003"],
    ]);
  });

  it("has no slots on a day without pickups", () => {
    expect(groupPickupsBySlot([])).toEqual([]);
  });
});

describe("isLatePickup", () => {
  const now = new Date("2026-10-01T09:30:00+07:00");

  it("flags an order still to collect once its slot has ended", () => {
    expect(isLatePickup(pickup(1, "2026-10-01T09:00:00+07:00", "ready"), 30, now)).toBe(true);
    expect(isLatePickup(pickup(1, "2026-10-01T09:00:00+07:00", "in_production"), 30, now)).toBe(true);
  });

  it("leaves a pickup within its slot, collected or missed alone", () => {
    expect(isLatePickup(pickup(1, "2026-10-01T09:00:00+07:00", "ready"), 60, now)).toBe(false);
    expect(isLatePickup(pickup(1, "2026-10-01T09:00:00+07:00", "fulfilled"), 30, now)).toBe(false);
    expect(isLatePickup(pickup(1, "2026-10-01T09:00:00+07:00", "no_show"), 30, now)).toBe(false);
  });
});
