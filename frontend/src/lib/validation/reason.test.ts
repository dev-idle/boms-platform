import { describe, expect, it } from "vitest";

import { reasonSchema } from "./reason";

describe("reasonSchema", () => {
  const schema = reasonSchema("Say why");

  it("keeps a plain reason, trimmed", () => {
    expect(schema.parse("  Out of matcha today  ")).toBe("Out of matcha today");
  });

  it("refuses an empty, long or formatted reason", () => {
    for (const reason of ["   ", "a".repeat(201), "Out of\nmatcha", "Zero\u200bwidth"]) {
      expect(schema.safeParse(reason).success).toBe(false);
    }
  });
});
