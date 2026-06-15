import {
  dashboardTableEmptyFilteredMessage,
  dashboardTableEmptyMessage,
  dashboardTableErrorMessage,
  dashboardTableLoadingMessage,
} from "@/constants/dashboard-table";

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
        <td className="db-table-empty-cell" colSpan={columnCount}>
          {dashboardTableLoadingMessage(entityLabel)}
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
