import { describe, expect, it } from "vitest";

import { closedDayBounds, formatClosedDay } from "./closed-days";

describe("closedDayBounds", () => {
  it("counts from the bakery's today, not the browser's", () => {
    // 18:00 UTC on the 12th is 01:00 on the 13th at the bakery (UTC+7).
    expect(closedDayBounds(new Date("2026-07-12T18:00:00Z"))).toEqual({
      min: "2026-07-13",
      max: "2027-07-13",
    });
  });
});

describe("formatClosedDay", () => {
  it("reads a calendar day without shifting it across zones", () => {
    expect(formatClosedDay("2026-10-20")).toBe("Tue, Oct 20, 2026");
  });
});
