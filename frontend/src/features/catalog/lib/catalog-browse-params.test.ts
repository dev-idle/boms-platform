import { describe, expect, it } from "vitest";

import { ROUTE } from "@/constants/routes";

import {
  buildCatalogBrowseHref,
  CATALOG_SEARCH_MAX_LENGTH,
  parseCatalogBrowseParams,
  toCatalogProductsFilter,
} from "./catalog-browse-params";

describe("CATALOG_SEARCH_MAX_LENGTH", () => {
  it("matches the browse params schema cap", () => {
    expect(CATALOG_SEARCH_MAX_LENGTH).toBe(100);
  });
});

describe("parseCatalogBrowseParams", () => {
  it("returns defaults for an empty query", () => {
    expect(parseCatalogBrowseParams(new URLSearchParams())).toEqual({
      search: "",
      category: undefined,
      page: 1,
    });
  });

  it("trims search and parses category uuid", () => {
    const categoryId = "550e8400-e29b-41d4-a716-446655440000";
    const params = parseCatalogBrowseParams(
      new URLSearchParams({
        search: "  croissant  ",
        category: categoryId,
        page: "3",
      }),
    );
    expect(params).toEqual({
      search: "croissant",
      category: categoryId,
      page: 3,
    });
  });

  it("falls back to page 1 for invalid page values", () => {
    expect(
      parseCatalogBrowseParams(new URLSearchParams({ page: "0" })).page,
    ).toBe(1);
  });
});

describe("buildCatalogBrowseHref", () => {
  it("omits empty filters from the query string", () => {
    expect(buildCatalogBrowseHref({ search: "", page: 1 })).toBe(ROUTE.products);
  });

  it("serializes active filters", () => {
    const categoryId = "550e8400-e29b-41d4-a716-446655440000";
    expect(
      buildCatalogBrowseHref({
        search: "cake",
        category: categoryId,
        page: 2,
      }),
    ).toBe(
      `${ROUTE.products}?search=cake&category=${categoryId}&page=2`,
    );
  });
});

describe("toCatalogProductsFilter", () => {
  it("maps browse params to the catalog products API filter", () => {
    const categoryId = "550e8400-e29b-41d4-a716-446655440000";
    expect(
      toCatalogProductsFilter(
        { search: "tart", category: categoryId, page: 2 },
        24,
      ),
    ).toEqual({
      page: 2,
      page_size: 24,
      category_id: categoryId,
      search: "tart",
    });
  });
});
