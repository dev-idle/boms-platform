"use client";

import { Button } from "@/components/ui/button";
import { REPORTED_INCIDENT_TYPES, type IncidentType } from "@/lib/schemas/incident";

import { useIncidentSummary } from "../hooks";

/** How many incidents of each type a week holds, none for a type left out. */
type IncidentCounts = Partial<Record<IncidentType, number>>;

const KPIS: ReadonlyArray<{ label: string; count: (counts: IncidentCounts) => number }> = [
  { label: "Incidents", count: (counts) => Object.values(counts).reduce((sum, count) => sum + count, 0) },
  {
    label: "Reported by staff",
    count: (counts) => REPORTED_INCIDENT_TYPES.reduce((sum, type) => sum + (counts[type] ?? 0), 0),
  },
  { label: "Ready late", count: (counts) => counts.ready_late ?? 0 },
  { label: "Payment flags", count: (counts) => counts.payment_anomaly ?? 0 },
];

type ManagerIncidentKpisProps = {
  week: string;
};

/**
 * A week's incidents in four numbers: all of them, the mistakes the counter
 * reported, orders ready after their pickup time, and customers flagged for
 * unusual payment activity.
 */
export function ManagerIncidentKpis({ week }: ManagerIncidentKpisProps) {
  const summaryQuery = useIncidentSummary(week);
  const counts: IncidentCounts | undefined = summaryQuery.data
    ? Object.fromEntries(summaryQuery.data.types.map(({ type, count }) => [type, count]))
    : undefined;

  if (summaryQuery.isError) {
    return (
      <div className="db-kpi-error">
        <p className="db-kpi-error__title">We could not count the week&apos;s incidents.</p>
        <Button type="button" variant="outline" onClick={() => void summaryQuery.refetch()}>
          Try again
        </Button>
      </div>
    );
  }

  return (
    <dl aria-busy={counts ? undefined : true} className="db-kpi-strip">
      {KPIS.map((kpi) => (
        <div className="db-kpi" key={kpi.label}>
          <dt className="db-kpi__label">{kpi.label}</dt>
          <dd className="db-kpi__value">
            {counts ? kpi.count(counts) : <span className="skeleton db-kpi__skeleton" />}
          </dd>
        </div>
      ))}
    </dl>
  );
}
