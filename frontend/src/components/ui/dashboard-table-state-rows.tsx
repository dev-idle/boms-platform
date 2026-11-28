import {
  dashboardTableEmptyFilteredMessage,
  dashboardTableEmptyMessage,
  dashboardTableErrorMessage,
} from "@/constants/dashboard-table";

import { Button } from "./button";

/** Rows the skeleton shows while a table's first page loads. */
const SKELETON_ROWS = 5;

type DashboardTableStateRowsProps = {
  columnCount: number;
  emptyFilteredMessage?: string;
  emptyMessage?: string;
  entityLabel: string;
  hasActiveFilter?: boolean;
  isEmpty: boolean;
  isError: boolean;
  initialLoading: boolean;
  /** Asks for the rows again after a failure. */
  onRetry: () => void;
};

/**
 * Loading, error and empty rows for dashboard tables: a skeleton of the rows to
 * come, a failure with the way to try again, or what the empty list means.
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
  onRetry,
}: DashboardTableStateRowsProps) {
  if (initialLoading) {
    return Array.from({ length: SKELETON_ROWS }, (_, row) => (
      <tr aria-hidden={row > 0 || undefined} key={row}>
        {Array.from({ length: columnCount }, (_, column) => (
          <td key={column}>
            {row === 0 && column === 0 ? <span className="sr-only" role="status">{`Loading ${entityLabel}`}</span> : null}
            <span className="skeleton block h-4 w-2/3" />
          </td>
        ))}
      </tr>
    ));
  }

  if (isError) {
    return (
      <tr>
        <td className="db-table-empty-cell" colSpan={columnCount}>
          <div className="db-table-error">
            <p className="text-error">{dashboardTableErrorMessage(entityLabel)}</p>
            <Button size="sm" type="button" variant="outline" onClick={onRetry}>
              Try again
            </Button>
          </div>
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
