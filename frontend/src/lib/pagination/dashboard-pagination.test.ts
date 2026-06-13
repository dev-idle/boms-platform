import { describe, expect, it } from "vitest";

import {
  buildPaginationWindow,
  formatPaginationRange,
  paginatedPlaceholderCount,
  paginatedPlaceholderCountFromMeta,
  paginationRange,
} from "./dashboard-pagination";

describe("paginationRange", () => {
  it("returns inclusive bounds for a middle page", () => {
    expect(paginationRange(3, 10, 127)).toEqual({ start: 21, end: 30 });
  });

  it("clamps the final page", () => {
    expect(paginationRange(13, 10, 127)).toEqual({ start: 121, end: 127 });
  });
});

describe("formatPaginationRange", () => {
  it("formats multi-row ranges", () => {
    expect(
      formatPaginationRange({ start: 11, end: 20 }, 127, "users"),
    ).toBe("Showing 11–20 of 127 users");
  });
});

describe("paginatedPlaceholderCount", () => {
  it("pads only partial pages when pagination spans multiple pages", () => {
    expect(paginatedPlaceholderCount(7, 10, 2)).toBe(3);
    expect(paginatedPlaceholderCount(3, 10, 1)).toBe(0);
    expect(paginatedPlaceholderCount(3, 10, 2)).toBe(7);
    expect(paginatedPlaceholderCount(10, 10, 2)).toBe(0);
    expect(paginatedPlaceholderCount(0, 10, 2)).toBe(0);
  });
});

describe("paginatedPlaceholderCountFromMeta", () => {
  it("uses pagination meta with a fallback page size", () => {
    expect(
      paginatedPlaceholderCountFromMeta(
        7,
        { page_size: 10, total_pages: 2 },
        20,
      ),
    ).toBe(3);
    expect(paginatedPlaceholderCountFromMeta(3, undefined, 10)).toBe(0);
  });
});

describe("buildPaginationWindow", () => {
  it("returns all pages for short lists", () => {
    expect(buildPaginationWindow(2, 4)).toEqual([1, 2, 3, 4]);
    expect(buildPaginationWindow(3, 7)).toEqual([1, 2, 3, 4, 5, 6, 7]);
  });

  it("shows a wide leading block on page 1", () => {
    expect(buildPaginationWindow(1, 13)).toEqual([
      1,
      2,
      3,
      4,
      5,
      6,
      "ellipsis",
      13,
    ]);
  });

  it("keeps the leading block through page 4", () => {
    expect(buildPaginationWindow(4, 13)).toEqual([
      1,
      2,
      3,
      4,
      5,
      6,
      "ellipsis",
      13,
    ]);
  });

  it("uses a centered window in the middle", () => {
    expect(buildPaginationWindow(6, 13)).toEqual([
      1,
      "ellipsis",
      4,
      5,
      6,
      7,
      8,
      "ellipsis",
      13,
    ]);
  });

  it("shows a wide trailing block near the end", () => {
    expect(buildPaginationWindow(13, 13)).toEqual([
      1,
      "ellipsis",
      8,
      9,
      10,
      11,
      12,
      13,
    ]);
  });
});
