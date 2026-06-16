import {
  dashboardTableEmptyFilteredMessage,
  dashboardTableEmptyMessage,
  dashboardTableErrorMessage,
} from "@/constants/dashboard-table";

import { DashboardBusyIndicator } from "./dashboard-busy-overlay";

type DashboardTableStateRowsProps = {
  columnCount: number;
  emptyFilteredMessage?: string;
  emptyMessage?: string;
  entityLabel: string;
  hasActiveFilter?: boolean;
  isEmpty: boolean;
  isError: boolean;
  initialLoading: boolean;
};

/**
 * Loading, error, and empty rows for dashboard tables.
 * Initial load: spinner in the first body row (row 2 under the header).
 */
export function DashboardTableStateRows({
  columnCount,
  emptyFilteredMessage,
  emptyMessage,
  entityLabel,
  hasActiveFilter = false,
  isEmpty,
  isError,
  initialLoading,
}: DashboardTableStateRowsProps) {
  if (initialLoading) {
    return (
      <tr>
        <td
          className="db-table-empty-cell db-table-loading-cell"
          colSpan={columnCount}
        >
          <div className="db-table-loading-indicator" role="status">
            <DashboardBusyIndicator />
          </div>
        </td>
      </tr>
    );
  }

  if (isError) {
    return (
      <tr>
        <td className="db-table-empty-cell text-error" colSpan={columnCount}>
          {dashboardTableErrorMessage(entityLabel)}
        </td>
      </tr>
    );
  }

  if (isEmpty) {
    const message = hasActiveFilter
      ? (emptyFilteredMessage ??
        dashboardTableEmptyFilteredMessage(entityLabel))
      : (emptyMessage ?? dashboardTableEmptyMessage(entityLabel));

    return (
      <tr>
        <td className="db-table-empty-cell" colSpan={columnCount}>
          {message}
        </td>
      </tr>
    );
  }

  return null;
}
