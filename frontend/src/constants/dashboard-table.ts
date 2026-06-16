/** Shared dashboard data-table defaults (admin + manager + staff lists). */
export const DASHBOARD_TABLE_PAGE_SIZE = 10;

/** Staff order queue shows more rows per page (operational throughput). */
export const DASHBOARD_STAFF_ORDERS_PAGE_SIZE = 20;

export function dashboardTableErrorMessage(entityLabel: string): string {
  return `Failed to load ${entityLabel}.`;
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
