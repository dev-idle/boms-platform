import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import {
  PHONE_FORMAT_MESSAGE,
  formatVietnamPhone,
  normalizeVietnamPhone,
  nullableVietnamPhoneZodString,
  vietnamPhoneZodString,
} from "./phone";

type PhoneCases = {
  normalize: { input: string; stored: string }[];
  reject: string[];
  display: { stored: string; display: string }[];
};

/** The shared fixture both the TypeScript and the Go rule are held to. */
const cases = JSON.parse(
  readFileSync(
    new URL("../../../../contracts/vietnam-phone-cases.json", import.meta.url),
    "utf8",
  ),
) as PhoneCases;

describe("normalizeVietnamPhone", () => {
  it("stores every accepted spelling in one form", () => {
    for (const { input, stored } of cases.normalize) {
      expect(normalizeVietnamPhone(input), input).toBe(stored);
    }
  });

  it("rejects what is not a Vietnam number", () => {
    for (const input of cases.reject) {
      expect(normalizeVietnamPhone(input), input).toBeNull();
    }
  });
});

describe("formatVietnamPhone", () => {
  it("groups mobiles 4-3-3 and splits land lines after the area code", () => {
    for (const { stored, display } of cases.display) {
      expect(formatVietnamPhone(stored), stored).toBe(display);
    }
  });

  it("is empty for no value", () => {
    expect(formatVietnamPhone(null)).toBe("");
    expect(formatVietnamPhone(undefined)).toBe("");
    expect(formatVietnamPhone("")).toBe("");
  });

  it("hands back anything it cannot read, so legacy rows still render", () => {
    expect(formatVietnamPhone("ext. 4412")).toBe("ext. 4412");
  });
});

describe("vietnamPhoneZodString", () => {
  const schema = vietnamPhoneZodString();

  it("normalizes a valid number", () => {
    expect(schema.parse(" 0912 345 678 ")).toBe("+84912345678");
  });

  it("keeps an empty field as the clear request and an omitted one as absent", () => {
    expect(schema.parse("  ")).toBe("");
    expect(schema.parse(undefined)).toBeUndefined();
  });

  it("rejects anything else with the format message", () => {
    const result = schema.safeParse("!!!!!!");
    expect(result.success).toBe(false);
    expect(result.error?.issues[0]?.message).toBe(PHONE_FORMAT_MESSAGE);
  });
});

describe("nullableVietnamPhoneZodString", () => {
  const schema = nullableVietnamPhoneZodString();

  it("maps an empty field to null", () => {
    expect(schema.parse("")).toBeNull();
    expect(schema.parse(null)).toBeNull();
  });

  it("normalizes and rejects like the PATCH field", () => {
    expect(schema.parse("+84 28 3822 1234")).toBe("+842838221234");
    expect(schema.safeParse("0912").success).toBe(false);
  });
});
