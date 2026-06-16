import { formatDateTime } from "@/lib/validation/datetime";

type DashboardTableDateTimeCellProps = {
  iso: string;
};

/** Dashboard table cell — single timestamp, no wrap. */
export function DashboardTableDateTimeCell({ iso }: DashboardTableDateTimeCellProps) {
  const label = formatDateTime(iso);

  return (
    <td className="db-table-datetime text-muted" title={label}>
      <time dateTime={iso}>{label}</time>
    </td>
  );
}
