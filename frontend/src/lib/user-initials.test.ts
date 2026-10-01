import { describe, expect, it } from "vitest";

import { initialsOf } from "./user-initials";

describe("initialsOf", () => {
  it("takes the first letters of a name's first two words", () => {
    expect(initialsOf("  linh  tran nguyen ", "linh@example.com")).toBe("LT");
  });

  it("takes the first two letters of a one-word name", () => {
    expect(initialsOf("Mai", "mai@example.com")).toBe("MA");
  });

  it("falls back to the start of the email without a name", () => {
    expect(initialsOf(null, "nghia@example.com")).toBe("NG");
    expect(initialsOf("   ", "nghia@example.com")).toBe("NG");
  });
});
