import { describe, expect, it } from "vitest";

import { initialOf } from "./user-initials";

describe("initialOf", () => {
  it("takes the first letter of the name", () => {
    expect(initialOf("  linh tran ", "linh@example.com")).toBe("L");
  });

  it("falls back to the email without a name", () => {
    expect(initialOf(null, "nghia@example.com")).toBe("N");
    expect(initialOf("   ", "nghia@example.com")).toBe("N");
  });
});
