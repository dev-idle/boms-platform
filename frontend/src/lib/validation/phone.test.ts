import { readFileSync } from "node:fs";

import { describe, expect, it } from "vitest";

import {
  PHONE_FORMAT_MESSAGE,
  PHONE_LENGTH_MESSAGE,
  formatNationalNumber,
  formatVietnamPhone,
  nationalNumber,
  normalizeVietnamPhone,
  phoneFormatMessage,
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

  it("rejects what is not a Vietnam mobile number", () => {
    for (const input of cases.reject) {
      expect(normalizeVietnamPhone(input), input).toBeNull();
    }
  });
});

describe("nationalNumber", () => {
  it("drops a leading 0 as it is typed", () => {
    expect(nationalNumber("0")).toBe("");
    expect(nationalNumber("09")).toBe("9");
    expect(nationalNumber("0912 345 678")).toBe("912345678");
  });

  it("drops a pasted +84, 84 or 0084 but keeps a national 84x prefix", () => {
    expect(nationalNumber("+84 912 345 678")).toBe("912345678");
    expect(nationalNumber("84912345678")).toBe("912345678");
    expect(nationalNumber("0084 912 345 678")).toBe("912345678");
    expect(nationalNumber("841234567")).toBe("841234567");
  });

  it("keeps zeros inside the number", () => {
    expect(nationalNumber("0900 001 234")).toBe("900001234");
  });
});

describe("formatNationalNumber", () => {
  it("groups digits in threes as they are typed", () => {
    expect(formatNationalNumber("")).toBe("");
    expect(formatNationalNumber("912")).toBe("912");
    expect(formatNationalNumber("9123")).toBe("912 3");
    expect(formatNationalNumber("912345678")).toBe("912 345 678");
  });
});

describe("formatVietnamPhone", () => {
  it("groups the national number after +84", () => {
    for (const { stored, display } of cases.display) {
      expect(formatVietnamPhone(stored), stored).toBe(display);
    }
  });

  it("is empty for no value and hands back what it cannot read", () => {
    expect(formatVietnamPhone(null)).toBe("");
    expect(formatVietnamPhone("")).toBe("");
    expect(formatVietnamPhone("ext. 4412")).toBe("ext. 4412");
  });
});

describe("vietnamPhoneZodString", () => {
  const schema = vietnamPhoneZodString();

  it("sends the national number as E.164", () => {
    expect(schema.parse("912345678")).toBe("+84912345678");
  });

  it("keeps an empty field as the clear request and an omitted one as absent", () => {
    expect(schema.parse("  ")).toBe("");
    expect(schema.parse(undefined)).toBeUndefined();
  });

  it("rejects anything else, saying what to fix", () => {
    const result = schema.safeParse("91234");
    expect(result.success).toBe(false);
    expect(result.error?.issues[0]?.message).toBe(PHONE_LENGTH_MESSAGE);
  });
});

describe("phoneFormatMessage", () => {
  it("names a prefix no carrier uses, however many digits follow", () => {
    expect(phoneFormatMessage("12345678")).toBe(
      "No Vietnam mobile number starts with 012",
    );
    expect(phoneFormatMessage("953456789")).toBe(
      "No Vietnam mobile number starts with 095",
    );
    expect(phoneFormatMessage("2")).toBe("No Vietnam mobile number starts with 02");
  });

  it("asks for the length when the prefix can still be right", () => {
    expect(phoneFormatMessage("9")).toBe(PHONE_LENGTH_MESSAGE);
    expect(phoneFormatMessage("91234")).toBe(PHONE_LENGTH_MESSAGE);
    expect(phoneFormatMessage("0912 345 6789")).toBe(PHONE_LENGTH_MESSAGE);
  });

  it("falls back to the format message for text that is not a phone", () => {
    expect(phoneFormatMessage("ext. 4412")).toBe(PHONE_FORMAT_MESSAGE);
  });
});
