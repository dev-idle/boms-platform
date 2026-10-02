"use client";

import { Button } from "@/components/ui/button";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useSalesReport } from "../hooks";
import type { SalesReport, SalesReportFilterInput } from "../schemas";

const KPIS: ReadonlyArray<{ label: string; money?: boolean; value: (report: SalesReport) => string }> = [
  { label: "Net revenue", money: true, value: (report) => formatPriceCents(report.totals.net_cents) },
  { label: "Orders", value: (report) => String(report.totals.orders) },
  {
    label: "Average order",
    money: true,
    value: (report) =>
      report.totals.average_order_cents === null ? "—" : formatPriceCents(report.totals.average_order_cents),
  },
  {
    label: "Production time",
    value: (report) =>
      report.production.average_minutes === null ? "—" : `${report.production.average_minutes} min`,
  },
];

type ManagerSalesKpisProps = {
  range: SalesReportFilterInput;
};

/** The range in four numbers: money kept after refunds, orders paid, their average, and how long orders took to make. */
export function ManagerSalesKpis({ range }: ManagerSalesKpisProps) {
  const reportQuery = useSalesReport(range);
  const report = reportQuery.data;

  if (reportQuery.isError) {
    return (
      <div className="db-kpi-error">
        <p className="db-kpi-error__title">We could not add up the sales.</p>
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
          <dd className={kpi.money ? "db-kpi__value db-kpi__value--money" : "db-kpi__value"}>
            {report ? kpi.value(report) : <span className="skeleton db-kpi__skeleton" />}
          </dd>
        </div>
      ))}
    </dl>
  );
}
