import { describe, expect, it } from "vitest";

import {
  meetsPasswordComplexity,
  newPasswordZodString,
  PASSWORD_REQUIREMENT_CHECKS,
} from "./password";

describe("meetsPasswordComplexity", () => {
  it("accepts lowercase letter with number", () => {
    expect(meetsPasswordComplexity("rose2024")).toBe(true);
  });

  it("accepts uppercase letter with number", () => {
    expect(meetsPasswordComplexity("Rose2024")).toBe(true);
  });

  it("rejects letters only", () => {
    expect(meetsPasswordComplexity("rosebuds")).toBe(false);
  });

  it("rejects numbers only", () => {
    expect(meetsPasswordComplexity("12345678")).toBe(false);
  });
});

describe("newPasswordZodString", () => {
  const schema = newPasswordZodString();

  it("rejects passwords that fail complexity", () => {
    const result = schema.safeParse("rosebuds");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(
        "Password must include a letter and a number",
      );
    }
  });

  it("accepts passwords that meet all rules", () => {
    expect(schema.safeParse("rose2024").success).toBe(true);
  });
});

describe("PASSWORD_REQUIREMENT_CHECKS", () => {
  it("uses the same letter and digit rules as meetsPasswordComplexity", () => {
    const letter = PASSWORD_REQUIREMENT_CHECKS.find((c) => c.id === "letter");
    const digit = PASSWORD_REQUIREMENT_CHECKS.find((c) => c.id === "digit");
    expect(letter).toBeDefined();
    expect(digit).toBeDefined();

    const password = "rose2024";
    expect(letter?.test(password)).toBe(true);
    expect(digit?.test(password)).toBe(true);
    expect(meetsPasswordComplexity(password)).toBe(true);
  });
});
