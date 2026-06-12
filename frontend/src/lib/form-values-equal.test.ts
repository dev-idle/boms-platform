import { describe, expect, it } from "vitest";

import { shallowFormValuesEqual } from "./form-values-equal";

describe("shallowFormValuesEqual", () => {
  it("compares all baseline keys strictly", () => {
    expect(
      shallowFormValuesEqual(
        { a: "1", b: "2" },
        { a: "1", b: "2" },
      ),
    ).toBe(true);
    expect(
      shallowFormValuesEqual(
        { a: "1", b: "3" },
        { a: "1", b: "2" },
      ),
    ).toBe(false);
  });

  it("treats null and empty string as equal for nullable keys", () => {
    expect(
      shallowFormValuesEqual(
        { phone: null },
        { phone: "" },
        { nullableKeys: ["phone"] },
      ),
    ).toBe(true);
  });
});
