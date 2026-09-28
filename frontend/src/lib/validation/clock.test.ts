import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import { clockTimeSchema, clockToMinutes, formatClockMinutes } from "./clock";

type ClockCases = {
  clock: {
    valid: { text: string; minutes: number }[];
    invalid: string[];
  };
};

/** The shared fixture both the TypeScript and the Go rule are held to. */
const cases = JSON.parse(
  readFileSync(
    new URL("../../../../contracts/pickup-rules-cases.json", import.meta.url),
    "utf8",
  ),
) as ClockCases;

describe("clockTimeSchema", () => {
  it("accepts every time of day and reads it as minutes after midnight", () => {
    for (const { text, minutes } of cases.clock.valid) {
      expect(clockTimeSchema.safeParse(text).success, text).toBe(true);
      expect(clockToMinutes(text), text).toBe(minutes);
    }
  });

  it("rejects anything else", () => {
    for (const text of cases.clock.invalid) {
      expect(clockTimeSchema.safeParse(text).success, text).toBe(false);
    }
  });
});

describe("formatClockMinutes", () => {
  it("writes minutes as a readable time", () => {
    expect(formatClockMinutes(0)).toBe("12:00 AM");
    expect(formatClockMinutes(480)).toBe("8:00 AM");
    expect(formatClockMinutes(1080)).toBe("6:00 PM");
    expect(formatClockMinutes(1439)).toBe("11:59 PM");
  });
});
