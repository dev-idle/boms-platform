import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import { clockToMinutes } from "./clock";
import {
  bakeryPickupISOFromLocalInput,
  defaultPickupLocalInputValue,
  formatPickupLocalInputValue,
  maxPickupLocalInputValue,
  pickupInstantFromLocalInput,
  pickupProblem,
  type PickupProblem,
  type PickupWindow,
} from "./pickup";

type PickupRulesCases = {
  policy: {
    rules: {
      opens_at: string;
      closes_at: string;
      preorder_min_lead_minutes: number;
      max_advance_days: number;
      closed_dates: string[];
    };
    now: string;
    cases: { at: string; problem: PickupProblem | null }[];
  };
};

/** The shared fixture both the TypeScript and the Go rule are held to. */
const cases = JSON.parse(
  readFileSync(
    new URL("../../../../contracts/pickup-rules-cases.json", import.meta.url),
    "utf8",
  ),
) as PickupRulesCases;

const WINDOW: PickupWindow = {
  opensAtMinutes: 8 * 60,
  closesAtMinutes: 18 * 60,
  minLeadMinutes: 120,
  maxAdvanceDays: 14,
  closedDates: new Map([["2026-07-13", "Inventory"]]),
};

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
    expect(defaultPickupLocalInputValue(WINDOW, now)).toBe("2026-07-10T12:00");
  });

  it("rounds up to the next five-minute mark", () => {
    expect(defaultPickupLocalInputValue(WINDOW, new Date("2026-07-10T03:00:20.000Z"))).toBe(
      "2026-07-10T12:05",
    );
    expect(defaultPickupLocalInputValue(WINDOW, new Date("2026-07-10T03:03:00.000Z"))).toBe(
      "2026-07-10T12:05",
    );
  });

  it("snaps forward to opening when earliest lands before open", () => {
    // 22:30 UTC = 05:30 bakery time next day; +2h = 07:30 → snap to 08:00.
    const now = new Date("2026-07-10T22:30:00.000Z");
    expect(defaultPickupLocalInputValue(WINDOW, now)).toBe("2026-07-11T08:00");
  });

  it("rolls to next day opening when earliest lands after close", () => {
    // 10:30 UTC = 17:30 bakery time; +2h = 19:30 (closed) → next day 08:00.
    const now = new Date("2026-07-10T10:30:00.000Z");
    expect(defaultPickupLocalInputValue(WINDOW, now)).toBe("2026-07-11T08:00");
  });

  it("skips closed days", () => {
    // 19:30 bakery time on the 12th → the 13th is closed → the 14th at opening.
    const now = new Date("2026-07-12T12:30:00.000Z");
    expect(defaultPickupLocalInputValue(WINDOW, now)).toBe("2026-07-14T08:00");
  });

  it("is empty when opening falls past the booking window", () => {
    // 09:00 bakery time; 23 hours' notice lands at 08:00 tomorrow, before a
    // 10:00 opening that is itself past the one-day window.
    const now = new Date("2026-07-10T02:00:00.000Z");
    const tight = { ...WINDOW, opensAtMinutes: 600, minLeadMinutes: 23 * 60, maxAdvanceDays: 1 };
    expect(defaultPickupLocalInputValue(tight, now)).toBe("");
  });

  it("is empty when the window holds no open time", () => {
    const now = new Date("2026-07-12T12:30:00.000Z");
    const shut = { ...WINDOW, maxAdvanceDays: 1 };
    expect(defaultPickupLocalInputValue(shut, now)).toBe("");
  });
});

describe("maxPickupLocalInputValue", () => {
  it("returns now + the booking window in bakery time", () => {
    const now = new Date("2026-07-10T03:00:00.000Z");
    expect(maxPickupLocalInputValue(WINDOW, now)).toBe("2026-07-24T10:00");
  });
});

describe("pickupProblem", () => {
  const { rules, now, cases: policyCases } = cases.policy;
  const pickupWindow: PickupWindow = {
    opensAtMinutes: clockToMinutes(rules.opens_at),
    closesAtMinutes: clockToMinutes(rules.closes_at),
    minLeadMinutes: rules.preorder_min_lead_minutes,
    maxAdvanceDays: rules.max_advance_days,
    closedDates: new Map(rules.closed_dates.map((day) => [day, "Closed"])),
  };
  const nowInstant = pickupInstantFromLocalInput(now);

  it("agrees with the backend on every shared case", () => {
    expect(nowInstant).not.toBeNull();
    for (const { at, problem } of policyCases) {
      expect(pickupProblem(at, pickupWindow, nowInstant!), at).toBe(problem);
    }
  });

  it("treats unreadable input as missing", () => {
    expect(pickupProblem("garbage", pickupWindow, nowInstant!)).toBe("missing");
  });
});
