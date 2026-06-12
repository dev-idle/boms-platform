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

const PAGINATION_EDGE_BLOCK = 6;
const PAGINATION_EDGE_SIBLINGS = 2;

function pushPageRange(
  items: PaginationWindowItem[],
  from: number,
  to: number,
): void {
  for (let current = from; current <= to; current += 1) {
    items.push(current);
  }
}

/** Compact page window — wide leading/trailing blocks, ellipses in the middle. */
export function buildPaginationWindow(
  page: number,
  totalPages: number,
): PaginationWindowItem[] {
  const safeTotalPages = Math.max(totalPages, 1);
  const safePage = Math.min(Math.max(page, 1), safeTotalPages);

  if (safeTotalPages <= PAGINATION_EDGE_BLOCK + 1) {
    return Array.from({ length: safeTotalPages }, (_, index) => index + 1);
  }

  const items: PaginationWindowItem[] = [];
  const startBlockEnd = Math.min(PAGINATION_EDGE_BLOCK, safeTotalPages - 1);
  const endBlockStart = safeTotalPages - PAGINATION_EDGE_BLOCK + 1;
  const startZoneEnd = startBlockEnd - PAGINATION_EDGE_SIBLINGS;

  if (safePage <= startZoneEnd) {
    pushPageRange(items, 1, startBlockEnd);

    if (startBlockEnd < safeTotalPages - 1) {
      items.push("ellipsis");
    }

    items.push(safeTotalPages);
    return items;
  }

  if (safePage >= endBlockStart) {
    items.push(1);

    if (endBlockStart > 2) {
      items.push("ellipsis");
    }

    pushPageRange(items, endBlockStart, safeTotalPages);
    return items;
  }

  const left = Math.max(2, safePage - PAGINATION_EDGE_SIBLINGS);
  const right = Math.min(safeTotalPages - 1, safePage + PAGINATION_EDGE_SIBLINGS);

  items.push(1);

  if (left > 2) {
    items.push("ellipsis");
  }

  pushPageRange(items, left, right);

  if (right < safeTotalPages - 1) {
    items.push("ellipsis");
  }

  items.push(safeTotalPages);

  return items;
}
