import { describe, expect, it } from "vitest";

import {
  hour12To24,
  hour24To12,
  partsToIso,
  parseIsoToLocalParts,
  shiftDraftDays,
} from "./datetime-calendar";

describe("hour12 conversion", () => {
  it("maps midnight and noon", () => {
    expect(hour24To12(0)).toEqual({ hour12: 12, meridiem: "AM" });
    expect(hour24To12(12)).toEqual({ hour12: 12, meridiem: "PM" });
  });

  it("round-trips through 12-hour parts", () => {
    expect(hour12To24(3, "PM")).toBe(15);
    expect(hour12To24(12, "AM")).toBe(0);
    expect(hour12To24(1, "AM")).toBe(1);
  });
});

describe("partsToIso", () => {
  it("preserves local wall time in ISO", () => {
    const parts = parseIsoToLocalParts("2026-06-17T08:35:00.000Z");
    const local = new Date(partsToIso(parts));
    expect(local.getHours()).toBe(parts.hour);
    expect(local.getMinutes()).toBe(parts.minute);
  });
});

describe("shiftDraftDays", () => {
  const base = { year: 2026, month: 0, day: 31, hour: 9, minute: 30 };

  it("rolls into the next month and keeps the time", () => {
    expect(shiftDraftDays(base, 1)).toEqual({ year: 2026, month: 1, day: 1, hour: 9, minute: 30 });
  });

  it("rolls back across a year boundary", () => {
    expect(shiftDraftDays({ ...base, month: 0, day: 1 }, -7)).toEqual({
      year: 2025,
      month: 11,
      day: 25,
      hour: 9,
      minute: 30,
    });
  });
});
