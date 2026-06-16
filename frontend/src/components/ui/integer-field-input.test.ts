import { describe, expect, it } from "vitest";

import { parseIntegerFieldDraft } from "./integer-field-input";

describe("parseIntegerFieldDraft", () => {
  it("returns min for empty or minus-only input", () => {
    expect(parseIntegerFieldDraft("", 42, 0)).toBe(0);
    expect(parseIntegerFieldDraft("   ", 42, 0)).toBe(0);
    expect(parseIntegerFieldDraft("-", 42, 0)).toBe(0);
  });

  it("parses integers and clamps to min", () => {
    expect(parseIntegerFieldDraft("1250", 0, 0)).toBe(1250);
    expect(parseIntegerFieldDraft("-3", 9, 0)).toBe(0);
  });

  it("clamps to max when provided", () => {
    expect(parseIntegerFieldDraft("150", 50, 1, 100)).toBe(100);
    expect(parseIntegerFieldDraft("0", 50, 1, 100)).toBe(1);
  });

  it("falls back when input is not numeric", () => {
    expect(parseIntegerFieldDraft("abc", 42, 0)).toBe(42);
    expect(parseIntegerFieldDraft("12.5", 42, 0)).toBe(12);
  });
});
