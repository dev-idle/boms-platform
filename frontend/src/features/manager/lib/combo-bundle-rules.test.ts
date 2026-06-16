import { describe, expect, it } from "vitest";

import { isValidComboBundle } from "./combo-bundle-rules";

describe("isValidComboBundle", () => {
  it("accepts two products", () => {
    expect(
      isValidComboBundle([
        { quantity: 1 },
        { quantity: 1 },
      ]),
    ).toBe(true);
  });

  it("accepts one product with quantity at least 2", () => {
    expect(isValidComboBundle([{ quantity: 2 }])).toBe(true);
  });

  it("rejects a single product with quantity 1", () => {
    expect(isValidComboBundle([{ quantity: 1 }])).toBe(false);
  });

  it("rejects empty bundle", () => {
    expect(isValidComboBundle([])).toBe(false);
  });
});
