"use client";

import { Button } from "@/components/ui/button";

import { useEngagementReport } from "../hooks";
import { shareOf } from "../lib/engagement";
import type { EngagementReport } from "../schemas";

const KPIS: ReadonlyArray<{ label: string; value: (report: EngagementReport) => string }> = [
  { label: "Customers", value: (report) => String(report.customers) },
  { label: "Using two or more", value: (report) => shareOf(report.used_two_or_more, report.customers) },
  { label: "Using one", value: (report) => shareOf(report.used_one, report.customers) },
  { label: "Using none", value: (report) => shareOf(report.used_none, report.customers) },
];

/** The registered customers, and the share using two or more features, one, or none. */
export function ManagerEngagementKpis() {
  const reportQuery = useEngagementReport();
  const report = reportQuery.data;

  if (reportQuery.isError) {
    return (
      <div className="db-kpi-error">
        <p className="db-kpi-error__title">We could not count how customers use the features.</p>
        <Button type="button" variant="outline" onClick={() => void reportQuery.refetch()}>
          Try again
        </Button>
      </div>
    );
  }

  return (
    <dl aria-busy={report ? undefined : true} className="db-kpi-strip">
      {KPIS.map((kpi) => (
        <div className="db-kpi" key={kpi.label}>
          <dt className="db-kpi__label">{kpi.label}</dt>
          <dd className="db-kpi__value">{report ? kpi.value(report) : <span className="skeleton db-kpi__skeleton" />}</dd>
        </div>
      ))}
    </dl>
  );
}
