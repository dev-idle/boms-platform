import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import type { OrderType } from "@/lib/schemas/order";

import { clockToMinutes } from "./clock";
import {
  bakeryDayOf,
  bakeryPickupISOFromLocalInput,
  earliestPickupLocalValue,
  formatPickupLocalInputValue,
  lastPickupDay,
  pickupInstantFromLocalInput,
  pickupOrderType,
  pickupProblem,
  pickupSlotsOfDay,
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
      slot_minutes: number;
      instant_prep_minutes: number;
      closed_dates: string[];
    };
    now: string;
    cases: {
      at: string;
      kitchen: boolean;
      lead_minutes: number;
      sold_out_on?: string;
      problem: PickupProblem | null;
      type: OrderType | null;
    }[];
  };
};

/** The shared fixture both the TypeScript and the Go rule are held to. */
const cases = JSON.parse(
  readFileSync(
    new URL("../../../../contracts/pickup-rules-cases.json", import.meta.url),
    "utf8",
  ),
) as PickupRulesCases;

const { rules } = cases.policy;
const WINDOW: PickupWindow = {
  opensAtMinutes: clockToMinutes(rules.opens_at),
  closesAtMinutes: clockToMinutes(rules.closes_at),
  slotMinutes: rules.slot_minutes,
  preorderLeadMinutes: rules.preorder_min_lead_minutes,
  instantPrepMinutes: rules.instant_prep_minutes,
  maxAdvanceDays: rules.max_advance_days,
  closedDates: new Map(rules.closed_dates.map((day) => [day, "Closed"])),
};
const NOW = pickupInstantFromLocalInput(cases.policy.now)!;
const KITCHEN = { kitchen: true, leadMinutes: 0, soldOutOn: null };
const COUNTER = { kitchen: false, leadMinutes: 0, soldOutOn: null };

describe("formatPickupLocalInputValue", () => {
  it("formats an instant as bakery wall-clock time (UTC+7)", () => {
    expect(formatPickupLocalInputValue(new Date("2026-07-11T07:00:00.000Z"))).toBe("2026-07-11T14:00");
  });

  it("rolls the date across midnight in bakery time", () => {
    expect(formatPickupLocalInputValue(new Date("2026-07-11T18:30:00.000Z"))).toBe("2026-07-12T01:30");
    expect(bakeryDayOf(new Date("2026-07-11T18:30:00.000Z"))).toBe("2026-07-12");
  });
});

describe("bakeryPickupISOFromLocalInput", () => {
  it("serializes local input to RFC3339 with bakery offset", () => {
    expect(bakeryPickupISOFromLocalInput("2026-07-11T14:00")).toBe("2026-07-11T14:00:00+07:00");
  });
});

describe("pickupInstantFromLocalInput", () => {
  it("round-trips a local value to the correct UTC instant", () => {
    expect(pickupInstantFromLocalInput("2026-07-11T14:00")?.toISOString()).toBe("2026-07-11T07:00:00.000Z");
  });

  it("rejects malformed input", () => {
    expect(pickupInstantFromLocalInput("not-a-date")).toBeNull();
    expect(pickupInstantFromLocalInput("2026-07-11")).toBeNull();
  });
});

describe("pickupProblem and pickupOrderType", () => {
  it("agree with the backend on every shared case", () => {
    expect(NOW).not.toBeNull();
    for (const { at, kitchen, lead_minutes, sold_out_on, problem, type } of cases.policy.cases) {
      const items = { kitchen, leadMinutes: lead_minutes, soldOutOn: sold_out_on ?? null };
      expect(pickupProblem(at, WINDOW, NOW, items), at).toBe(problem);
      if (problem === null) {
        expect(pickupOrderType(at, NOW, items), at).toBe(type);
      }
    }
  });

  it("treats unreadable input as missing", () => {
    expect(pickupProblem("garbage", WINDOW, NOW, KITCHEN)).toBe("missing");
    expect(pickupOrderType("garbage", NOW, KITCHEN)).toBeNull();
  });
});

describe("pickupSlotsOfDay", () => {
  it("counts slots from opening, the last one starting before closing", () => {
    const window = { ...WINDOW, opensAtMinutes: 8 * 60 + 15, closesAtMinutes: 10 * 60 };
    expect(pickupSlotsOfDay("2026-07-10", window)).toEqual([
      "2026-07-10T08:15",
      "2026-07-10T08:45",
      "2026-07-10T09:15",
      "2026-07-10T09:45",
    ]);
  });
});

describe("lastPickupDay", () => {
  it("is the bakery day the booking window reaches", () => {
    expect(lastPickupDay(WINDOW, NOW)).toBe("2026-07-24");
  });
});

describe("earliestPickupLocalValue", () => {
  it("is the first slot after the pre-order notice", () => {
    expect(earliestPickupLocalValue(WINDOW, NOW, KITCHEN)).toBe("2026-07-10T12:00");
  });

  it("is sooner for counter items collected today", () => {
    expect(earliestPickupLocalValue(WINDOW, NOW, COUNTER)).toBe("2026-07-10T10:30");
  });

  it("waits the longest notice an item needs", () => {
    expect(earliestPickupLocalValue(WINDOW, NOW, { kitchen: true, leadMinutes: 24 * 60, soldOutOn: null })).toBe(
      "2026-07-11T10:00",
    );
  });

  it("skips the day an item ran out on", () => {
    expect(earliestPickupLocalValue(WINDOW, NOW, { ...COUNTER, soldOutOn: "2026-07-10" })).toBe("2026-07-11T08:00");
  });

  it("moves to the next open day after closing and past closed days", () => {
    const evening = pickupInstantFromLocalInput("2026-07-12T17:00")!;
    expect(earliestPickupLocalValue(WINDOW, evening, KITCHEN)).toBe("2026-07-14T08:00");
  });

  it("is empty when the booking window holds no open slot", () => {
    const evening = pickupInstantFromLocalInput("2026-07-12T17:00")!;
    expect(earliestPickupLocalValue({ ...WINDOW, maxAdvanceDays: 1 }, evening, KITCHEN)).toBe("");
  });
});
