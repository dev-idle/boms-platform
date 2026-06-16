import { describe, expect, it } from "vitest";

import {
  hour12To24,
  hour24To12,
  partsToIso,
  parseIsoToLocalParts,
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
