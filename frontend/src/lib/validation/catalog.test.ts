import { describe, expect, it } from "vitest";

import { catalogSlugSchema, slugifyCatalogName } from "./catalog";

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
