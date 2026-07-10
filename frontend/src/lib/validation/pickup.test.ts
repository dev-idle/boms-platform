import { describe, expect, it } from "vitest";

import {
  bakeryPickupISOFromLocalInput,
  defaultPickupLocalInputValue,
  formatPickupLocalInputValue,
  isPickupLocalValueValid,
  maxPickupLocalInputValue,
  pickupInstantFromLocalInput,
} from "./pickup";

describe("formatPickupLocalInputValue", () => {
  it("formats an instant as bakery wall-clock time (UTC+7)", () => {
    expect(
      formatPickupLocalInputValue(new Date("2026-07-11T07:00:00.000Z")),
    ).toBe("2026-07-11T14:00");
  });

  it("rolls the date across midnight in bakery time", () => {
    expect(
      formatPickupLocalInputValue(new Date("2026-07-11T18:30:00.000Z")),
    ).toBe("2026-07-12T01:30");
  });
});

describe("bakeryPickupISOFromLocalInput", () => {
  it("serializes local input to RFC3339 with bakery offset", () => {
    expect(bakeryPickupISOFromLocalInput("2026-07-11T14:00")).toBe(
      "2026-07-11T14:00:00+07:00",
    );
  });
});

describe("pickupInstantFromLocalInput", () => {
  it("round-trips a local value to the correct UTC instant", () => {
    const instant = pickupInstantFromLocalInput("2026-07-11T14:00");
    expect(instant?.toISOString()).toBe("2026-07-11T07:00:00.000Z");
  });

  it("rejects malformed input", () => {
    expect(pickupInstantFromLocalInput("not-a-date")).toBeNull();
    expect(pickupInstantFromLocalInput("2026-07-11")).toBeNull();
  });
});

describe("defaultPickupLocalInputValue", () => {
  it("returns now + lead time when inside business hours", () => {
    // 03:00 UTC = 10:00 bakery time; +2h lead = 12:00 (open).
    const now = new Date("2026-07-10T03:00:00.000Z");
    expect(defaultPickupLocalInputValue(now)).toBe("2026-07-10T12:00");
  });

  it("snaps forward to opening when earliest lands before open", () => {
    // 22:30 UTC = 05:30 bakery time next day; +2h = 07:30 → snap to 08:00.
    const now = new Date("2026-07-10T22:30:00.000Z");
    expect(defaultPickupLocalInputValue(now)).toBe("2026-07-11T08:00");
  });

  it("rolls to next day opening when earliest lands after close", () => {
    // 10:30 UTC = 17:30 bakery time; +2h = 19:30 (closed) → next day 08:00.
    const now = new Date("2026-07-10T10:30:00.000Z");
    expect(defaultPickupLocalInputValue(now)).toBe("2026-07-11T08:00");
  });
});

describe("maxPickupLocalInputValue", () => {
  it("returns now + 14 days in bakery time", () => {
    const now = new Date("2026-07-10T03:00:00.000Z");
    expect(maxPickupLocalInputValue(now)).toBe("2026-07-24T10:00");
  });
});

describe("isPickupLocalValueValid", () => {
  // 03:00 UTC = 10:00 bakery time.
  const now = new Date("2026-07-10T03:00:00.000Z");

  it("accepts a slot inside every constraint", () => {
    expect(isPickupLocalValueValid("2026-07-10T14:00", now)).toBe(true);
  });

  it("rejects a slot before the lead time", () => {
    expect(isPickupLocalValueValid("2026-07-10T11:59", now)).toBe(false);
  });

  it("accepts exactly at the lead-time boundary", () => {
    expect(isPickupLocalValueValid("2026-07-10T12:00", now)).toBe(true);
  });

  it("rejects a slot beyond the 14-day advance limit", () => {
    expect(isPickupLocalValueValid("2026-07-24T10:01", now)).toBe(false);
  });

  it("rejects slots outside business hours", () => {
    expect(isPickupLocalValueValid("2026-07-11T07:59", now)).toBe(false);
    expect(isPickupLocalValueValid("2026-07-11T18:00", now)).toBe(false);
  });

  it("accepts business-hour boundaries", () => {
    expect(isPickupLocalValueValid("2026-07-11T08:00", now)).toBe(true);
    expect(isPickupLocalValueValid("2026-07-11T17:59", now)).toBe(true);
  });

  it("rejects malformed input", () => {
    expect(isPickupLocalValueValid("", now)).toBe(false);
    expect(isPickupLocalValueValid("garbage", now)).toBe(false);
  });
});
