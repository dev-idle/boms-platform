import {
  dashboardTableEmptyFilteredMessage,
  dashboardTableEmptyMessage,
  dashboardTableErrorMessage,
} from "@/constants/dashboard-table";
import { LOADING_MESSAGE } from "@/constants/loading-copy";

import { LoadingIndicator } from "./loading-state";

type DashboardTableStateRowsProps = {
  columnCount: number;
  emptyFilteredMessage?: string;
  emptyMessage?: string;
  entityLabel: string;
  hasActiveFilter?: boolean;
  isEmpty: boolean;
  isError: boolean;
  isInitialLoad: boolean;
};

/** Loading, error, and empty rows for dashboard data tables. */
export function DashboardTableStateRows({
  columnCount,
  emptyFilteredMessage,
  emptyMessage,
  entityLabel,
  hasActiveFilter = false,
  isEmpty,
  isError,
  isInitialLoad,
}: DashboardTableStateRowsProps) {
  if (isInitialLoad) {
    return (
      <tr>
        <td className="db-table-empty-cell db-table-loading-cell" colSpan={columnCount}>
          <div className="db-table-loading-indicator" role="status">
            <span className="sr-only">{LOADING_MESSAGE}</span>
            <LoadingIndicator dots={6} />
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
