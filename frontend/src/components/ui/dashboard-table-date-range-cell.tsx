import { formatDateTime, splitDateTime } from "@/lib/validation/datetime";

type DashboardTableDateRangeCellProps = {
  startsAt: string;
  endsAt: string;
};

/**
 * Dashboard table cell — the days something runs from and to, one over the
 * other. The times of day are in its title: a list is scanned by day, and two
 * timestamp columns push a table's actions off a laptop screen.
 */
export function DashboardTableDateRangeCell({ startsAt, endsAt }: DashboardTableDateRangeCellProps) {
  return (
    <td className="db-table-datetime" title={`${formatDateTime(startsAt)} to ${formatDateTime(endsAt)}`}>
      <time dateTime={startsAt}>
        <span className="db-table-datetime__date">{splitDateTime(startsAt).date}</span>
      </time>
      <time dateTime={endsAt}>
        <span className="db-table-datetime__date">to {splitDateTime(endsAt).date}</span>
      </time>
    </td>
  );
}
