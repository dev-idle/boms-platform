import { describe, expect, it } from "vitest";

import {
  orderTypeNote,
  pickupHint,
  pickupProblemMessage,
  toPickupFulfillment,
  toPickupWindow,
} from "./pickup-window";

const pickupWindow = toPickupWindow({
  opens_at: "08:00",
  closes_at: "18:30",
  slot_minutes: 30,
  preorder_min_lead_minutes: 90,
  instant_prep_minutes: 20,
  max_advance_days: 14,
  closed_dates: [{ date: "2026-10-20", reason: "Staff training" }],
});
const cake = toPickupFulfillment({ has_kitchen_items: true, lead_minutes: 24 * 60, sold_out_on: null });
const pastries = toPickupFulfillment({ has_kitchen_items: false, lead_minutes: 0, sold_out_on: null });

describe("toPickupWindow", () => {
  it("reads the API rules in minutes and indexes closed days", () => {
    expect(pickupWindow.opensAtMinutes).toBe(480);
    expect(pickupWindow.closesAtMinutes).toBe(1110);
    expect(pickupWindow.slotMinutes).toBe(30);
    expect(pickupWindow.preorderLeadMinutes).toBe(90);
    expect(pickupWindow.instantPrepMinutes).toBe(20);
    expect(pickupWindow.closedDates.get("2026-10-20")).toBe("Staff training");
  });

  it("reads what the cart's items need", () => {
    expect(cake).toEqual({ kitchen: true, leadMinutes: 1440, soldOutOn: null });
  });
});

describe("pickup copy", () => {
  it("describes the window in force", () => {
    expect(pickupHint(pickupWindow)).toBe(
      "Pickup slots every 30 minutes, 8:00 AM–6:30 PM bakery time, up to 14 days ahead.",
    );
  });

  it("says how the order will be prepared and the notice it waits", () => {
    expect(orderTypeNote("instant", pickupWindow, pastries)).toBe(
      "Instant pickup: ready-made items, packed 20 minutes after you order.",
    );
    expect(orderTypeNote("pre_order", pickupWindow, pastries)).toBe("Pre-order: place it at least 1 hour 30 minutes before pickup.");
    expect(orderTypeNote("pre_order", pickupWindow, cake)).toBe("Pre-order: place it at least 24 hours before pickup.");
    expect(orderTypeNote("instant", { ...pickupWindow, instantPrepMinutes: 0 }, pastries)).toBe(
      "Instant pickup: ready-made items, ready as soon as you order.",
    );
  });

  it("names each problem the way checkout would", () => {
    expect(pickupProblemMessage("too_soon", pickupWindow, pastries, "pre_order")).toBe(
      "Pickup must be at least 1 hour 30 minutes from now.",
    );
    expect(pickupProblemMessage("too_soon", pickupWindow, pastries, "instant")).toBe(
      "Pickup must be at least 20 minutes from now.",
    );
    expect(pickupProblemMessage("too_far", pickupWindow, cake, "pre_order")).toBe(
      "Pickup can be at most 14 days ahead.",
    );
    expect(pickupProblemMessage("closed_day", pickupWindow, cake, "pre_order", "Staff training")).toBe(
      "The bakery is closed that day (Staff training). Choose another day.",
    );
    expect(pickupProblemMessage("outside_hours", pickupWindow, cake, "pre_order")).toBe(
      "Pickup must be between 8:00 AM–6:30 PM bakery time.",
    );
    expect(pickupProblemMessage("off_slot", pickupWindow, cake, "pre_order")).toBe("Choose one of the pickup slots.");
  });
});
