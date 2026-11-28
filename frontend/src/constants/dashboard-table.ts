/** Rows a page of any dashboard table holds, so no list outruns the screen. */
export const DASHBOARD_TABLE_PAGE_SIZE = 10;

export function dashboardTableErrorMessage(entityLabel: string): string {
  return `We could not load the ${entityLabel}.`;
}

export function dashboardTableEmptyMessage(entityLabel: string): string {
  return `No ${entityLabel} found.`;
}

export function dashboardTableEmptyFilteredMessage(
  entityLabel: string,
): string {
  return `No ${entityLabel} match your search.`;
}

export function dashboardTableEmptyFiltersMessage(
  entityLabel: string,
): string {
  return `No ${entityLabel} match your filters.`;
}
