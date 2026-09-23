import { formatDateTime, splitDateTime } from "@/lib/validation/datetime";

type DashboardTableDateTimeCellProps = {
  iso: string;
};

/** Dashboard table cell — the date, with its time on the line below. */
export function DashboardTableDateTimeCell({ iso }: DashboardTableDateTimeCellProps) {
  const { date, time } = splitDateTime(iso);

  return (
    <td className="db-table-datetime text-muted" title={formatDateTime(iso)}>
      <time dateTime={iso}>
        <span className="db-table-datetime__date">{date}</span>
        <span className="db-table-datetime__time">{time}</span>
      </time>
    </td>
  );
}
