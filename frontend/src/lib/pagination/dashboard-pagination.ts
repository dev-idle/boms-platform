export type PaginationRange = {
  end: number;
  start: number;
};

export type PaginationWindowItem = number | "ellipsis";

/** Inclusive row range for the current page (1-based indices). */
export function paginationRange(
  page: number,
  pageSize: number,
  totalItems: number,
): PaginationRange {
  if (totalItems <= 0) {
    return { end: 0, start: 0 };
  }

  const safePage = Math.max(page, 1);
  const start = (safePage - 1) * pageSize + 1;
  const end = Math.min(safePage * pageSize, totalItems);

  return { end, start };
}

export function formatPaginationRange(
  range: PaginationRange,
  totalItems: number,
  itemLabel: string,
): string {
  if (totalItems <= 0) {
    return `No ${itemLabel}`;
  }

  if (range.start === range.end) {
    return `Showing ${range.start} of ${totalItems} ${itemLabel}`;
  }

  return `Showing ${range.start}–${range.end} of ${totalItems} ${itemLabel}`;
}

/** Pad only a partial page in a multi-page list so height matches page 1. */
export function paginatedPlaceholderCount(
  itemCount: number,
  pageSize: number,
  totalPages: number,
): number {
  if (totalPages <= 1 || itemCount <= 0 || itemCount >= pageSize) {
    return 0;
  }

  return pageSize - itemCount;
}

type PaginatedPlaceholderMeta = {
  page_size: number;
  total_pages: number;
};

/** Resolve placeholder count from API pagination meta and a fallback page size. */
export function paginatedPlaceholderCountFromMeta(
  itemCount: number,
  pagination: PaginatedPlaceholderMeta | undefined,
  fallbackPageSize: number,
): number {
  return paginatedPlaceholderCount(
    itemCount,
    pagination?.page_size ?? fallbackPageSize,
    pagination?.total_pages ?? 1,
  );
}

/** Page slots the pager shows once a list has more pages than that. */
const PAGINATION_SLOTS = 7;

/**
 * The pages the pager offers: every page of a short list; past seven, always
 * seven slots — the first and last page, the current one with a neighbour on
 * each side, and an ellipsis for each gap — so the pager keeps its width and
 * its buttons their places as the reader pages through.
 */
export function buildPaginationWindow(
  page: number,
  totalPages: number,
): PaginationWindowItem[] {
  const last = Math.max(totalPages, 1);
  const current = Math.min(Math.max(page, 1), last);

  if (last <= PAGINATION_SLOTS) {
    return Array.from({ length: last }, (_, index) => index + 1);
  }
  if (current <= 4) {
    return [1, 2, 3, 4, 5, "ellipsis", last];
  }
  if (current >= last - 3) {
    return [1, "ellipsis", last - 4, last - 3, last - 2, last - 1, last];
  }
  return [1, "ellipsis", current - 1, current, current + 1, "ellipsis", last];
}

/**
 * Steps a table back to its last page when the list shrank under the page on
 * show — a row deleted, a filter that empties its last page — instead of
 * leaving an empty table with no pagination to leave it by. Called while
 * rendering, with the page count of the settled response.
 */
export function keepPageInRange(page: number, setPage: (page: number) => void, totalPages: number | undefined): void {
  const lastPage = Math.max(totalPages ?? page, 1);
  if (page > lastPage) {
    setPage(lastPage);
  }
}
