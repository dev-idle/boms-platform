import { describe, expect, it } from "vitest";

import { pickupHint, pickupProblemMessage, toPickupWindow } from "./pickup-window";

const pickupWindow = toPickupWindow({
  opens_at: "08:00",
  closes_at: "18:30",
  preorder_min_lead_minutes: 90,
  max_advance_days: 14,
  closed_dates: [{ date: "2026-10-20", reason: "Staff training" }],
});

describe("toPickupWindow", () => {
  it("reads the API rules in minutes and indexes closed days", () => {
    expect(pickupWindow.opensAtMinutes).toBe(480);
    expect(pickupWindow.closesAtMinutes).toBe(1110);
    expect(pickupWindow.minLeadMinutes).toBe(90);
    expect(pickupWindow.closedDates.get("2026-10-20")).toBe("Staff training");
  });
});

describe("pickup copy", () => {
  it("describes the window in force", () => {
    expect(pickupHint(pickupWindow)).toBe(
      "Collect your order at the bakery between 8:00 AM–6:30 PM bakery time, " +
        "at least 1 hour 30 minutes from now and up to 14 days ahead.",
    );
  });

  it("names each problem the way checkout would", () => {
    expect(pickupProblemMessage("too_soon", pickupWindow)).toBe(
      "Pickup must be at least 1 hour 30 minutes from now.",
    );
    expect(pickupProblemMessage("too_far", pickupWindow)).toBe("Pickup can be at most 14 days ahead.");
    expect(pickupProblemMessage("closed_day", pickupWindow, "Staff training")).toBe(
      "The bakery is closed that day (Staff training). Choose another day.",
    );
    expect(pickupProblemMessage("outside_hours", pickupWindow)).toBe(
      "Pickup must be between 8:00 AM–6:30 PM bakery time.",
    );
  });

  it("speaks of a lead time under an hour in minutes", () => {
    const quick = { ...pickupWindow, minLeadMinutes: 45 };
    expect(pickupProblemMessage("too_soon", quick)).toBe("Pickup must be at least 45 minutes from now.");
  });
});
