import { describe, expect, it } from "vitest";

import {
  CATALOG_COMBOS_HEADING_ID,
  CATALOG_COMBOS_SCROLL_ANCHOR_ID,
  CATALOG_COMBOS_SECTION_ID,
  resolveStorefrontCombosHeadingHash,
} from "./scroll-to-storefront-anchor";

describe("resolveStorefrontCombosHeadingHash", () => {
  it("returns the header id for the combo title hash", () => {
    expect(resolveStorefrontCombosHeadingHash(CATALOG_COMBOS_HEADING_ID)).toBe(
      CATALOG_COMBOS_HEADING_ID,
    );
  });

  it("maps the legacy scroll hash to the header id", () => {
    expect(resolveStorefrontCombosHeadingHash(CATALOG_COMBOS_SCROLL_ANCHOR_ID)).toBe(
      CATALOG_COMBOS_HEADING_ID,
    );
  });

  it("maps the legacy section hash to the header id", () => {
    expect(resolveStorefrontCombosHeadingHash(CATALOG_COMBOS_SECTION_ID)).toBe(
      CATALOG_COMBOS_HEADING_ID,
    );
  });

  it("returns null for unrelated hashes", () => {
    expect(resolveStorefrontCombosHeadingHash("products")).toBeNull();
  });

  it("returns null when there is no hash", () => {
    expect(resolveStorefrontCombosHeadingHash("")).toBeNull();
  });
});
