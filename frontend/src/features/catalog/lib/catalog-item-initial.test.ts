import { describe, expect, it } from "vitest";

import { catalogItemInitial } from "./catalog-item-initial";

describe("catalogItemInitial", () => {
  it("returns a single uppercase letter", () => {
    expect(catalogItemInitial("Butter Cookie Tin")).toBe("B");
    expect(catalogItemInitial("tiramisu cup")).toBe("T");
  });

  it("ignores leading whitespace", () => {
    expect(catalogItemInitial("   madeleine")).toBe("M");
  });

  it("falls back to an em dash when there is no letter", () => {
    expect(catalogItemInitial("   ")).toBe("—");
    expect(catalogItemInitial("")).toBe("—");
  });
});
