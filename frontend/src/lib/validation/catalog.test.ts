import { describe, expect, it } from "vitest";

import {
  acceptDollarDraft,
  catalogSlugSchema,
  formatDollarInput,
  parseDollarInput,
  slugifyCatalogName,
} from "./catalog";

describe("slugifyCatalogName", () => {
  it("lowercases and hyphenates spaces", () => {
    expect(slugifyCatalogName("Matcha Drinks")).toBe("matcha-drinks");
  });

  it("strips diacritics to match backend SlugFromName", () => {
    expect(slugifyCatalogName("Caf\u00e9")).toBe("cafe");
  });

  it("strips invalid characters", () => {
    expect(slugifyCatalogName("  Breads & Pastries!  ")).toBe("breads-pastries");
  });

  it("collapses repeated hyphens", () => {
    expect(slugifyCatalogName("a--b")).toBe("a-b");
  });

  it("returns empty for names with no slug characters", () => {
    expect(slugifyCatalogName("!!!")).toBe("");
  });

  it("trims trailing hyphens after 128-char truncation", () => {
    const longName = `${"a".repeat(127)}-extra`;
    const slug = slugifyCatalogName(longName);
    expect(slug.length).toBeLessThanOrEqual(128);
    expect(slug.endsWith("-")).toBe(false);
  });

  it("output passes catalogSlugSchema when non-empty", () => {
    const slug = slugifyCatalogName("Sourdough Loaf");
    expect(catalogSlugSchema.safeParse(slug).success).toBe(true);
  });
});

describe("parseDollarInput", () => {
  it("reads whole dollars and cents exactly", () => {
    expect(parseDollarInput("12")).toBe(1200);
    expect(parseDollarInput("12.50")).toBe(1250);
    expect(parseDollarInput("0.07")).toBe(7);
    expect(parseDollarInput("4.5")).toBe(450);
    expect(parseDollarInput("0")).toBe(0);
  });

  it("does not inherit float rounding", () => {
    // 4.50 * 100 is 450.00000000000006 in JavaScript; the digits are read instead.
    for (const [input, cents] of [
      ["4.50", 450],
      ["1.10", 110],
      ["0.29", 29],
      ["8.70", 870],
    ] as const) {
      expect(parseDollarInput(input), input).toBe(cents);
    }
  });

  it("accepts what a person types around the number", () => {
    expect(parseDollarInput("  12.50 ")).toBe(1250);
    expect(parseDollarInput("$12.50")).toBe(1250);
    expect(parseDollarInput("1,200")).toBe(120_000);
  });

  it("rejects anything that is not an amount", () => {
    for (const raw of ["", " ", ".", "-5", "12.505", "12.5.0", "abc", "1e3", "12 50"]) {
      expect(parseDollarInput(raw), raw).toBeNull();
    }
  });
});

describe("acceptDollarDraft", () => {
  it("lets an amount be typed one character at a time", () => {
    for (const raw of ["", "1", "12", "12.", "12.5", "12.50", "0", "0.07"]) {
      expect(acceptDollarDraft(raw), raw).toBe(raw);
    }
  });

  it("cleans a pasted amount instead of refusing it", () => {
    expect(acceptDollarDraft("$12.50")).toBe("12.50");
    expect(acceptDollarDraft("1,200")).toBe("1200");
  });

  it("refuses a keystroke that would not leave an amount", () => {
    // `.5` belongs here: the parser needs a leading digit, so a draft that
    // accepted it would look filled and commit nothing.
    for (const raw of ["0.00a", "abc", "-5", "12.505", "12.5.0", "1e3", "12 50", ".5"]) {
      expect(acceptDollarDraft(raw), raw).toBeNull();
    }
  });
});

describe("formatDollarInput", () => {
  it("always shows two decimals", () => {
    expect(formatDollarInput(1250)).toBe("12.50");
    expect(formatDollarInput(1200)).toBe("12.00");
    expect(formatDollarInput(7)).toBe("0.07");
    expect(formatDollarInput(0)).toBe("0.00");
  });

  it("round-trips through the parser", () => {
    for (const cents of [0, 1, 99, 100, 450, 120_000]) {
      expect(parseDollarInput(formatDollarInput(cents)), String(cents)).toBe(cents);
    }
  });
});
