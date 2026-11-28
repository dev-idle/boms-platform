import { describe, expect, it, vi } from "vitest";

import {
  buildPaginationWindow,
  formatPaginationRange,
  keepPageInRange,
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

  it("keeps the first five pages in view near the start", () => {
    expect(buildPaginationWindow(1, 13)).toEqual([1, 2, 3, 4, 5, "ellipsis", 13]);
    expect(buildPaginationWindow(4, 13)).toEqual([1, 2, 3, 4, 5, "ellipsis", 13]);
  });

  it("centres the current page in the middle", () => {
    expect(buildPaginationWindow(5, 13)).toEqual([1, "ellipsis", 4, 5, 6, "ellipsis", 13]);
    expect(buildPaginationWindow(9, 13)).toEqual([1, "ellipsis", 8, 9, 10, "ellipsis", 13]);
  });

  it("keeps the last five pages in view near the end", () => {
    expect(buildPaginationWindow(10, 13)).toEqual([1, "ellipsis", 9, 10, 11, 12, 13]);
    expect(buildPaginationWindow(13, 13)).toEqual([1, "ellipsis", 9, 10, 11, 12, 13]);
  });

  it("shows seven slots on every page of a long list", () => {
    for (let page = 1; page <= 40; page += 1) {
      expect(buildPaginationWindow(page, 40)).toHaveLength(7);
    }
  });
});

describe("keepPageInRange", () => {
  it("steps back to the last page when the list shrank under the page on show", () => {
    const setPage = vi.fn();
    keepPageInRange(3, setPage, 2);
    expect(setPage).toHaveBeenCalledWith(2);
  });

  it("goes to the first page when nothing is left", () => {
    const setPage = vi.fn();
    keepPageInRange(2, setPage, 0);
    expect(setPage).toHaveBeenCalledWith(1);
  });

  it("leaves a page inside the list, or one whose count is not known yet, alone", () => {
    const setPage = vi.fn();
    keepPageInRange(2, setPage, 2);
    keepPageInRange(4, setPage, undefined);
    expect(setPage).not.toHaveBeenCalled();
  });
});
