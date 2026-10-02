"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { Button } from "@/components/ui/button";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { saveFile } from "@/lib/dom/save-file";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useSalesReport } from "../hooks";
import { centsToAmount, downloadBlockedReason, formatPeriod, toCsv } from "../lib/sales-report";
import type { SalesReport, SalesReportFilterInput } from "../schemas";
import { ManagerSalesChart } from "./manager-sales-chart";

const COLUMN_COUNT = 5;

/** The report's periods as a spreadsheet: amounts in dollars, as the table shows them. */
function savePeriods(report: SalesReport): void {
  const rows = [
    ["Period", "Orders", "Sales (USD)", "Refunds (USD)", "Net (USD)"],
    ...report.periods.map((period) => [
      period.start,
      period.orders,
      centsToAmount(period.gross_cents),
      centsToAmount(period.refunds_cents),
      centsToAmount(period.net_cents),
    ]),
  ];
  saveFile(`sales-by-${report.group}-${report.from}-to-${report.to}.csv`, toCsv(rows), "text/csv");
}

type ManagerSalesPeriodsProps = {
  range: SalesReportFilterInput;
};

/** Net revenue by period, drawn and listed with what each period took and gave back. */
export function ManagerSalesPeriods({ range }: ManagerSalesPeriodsProps) {
  const reportQuery = useSalesReport(range);
  const { initialLoading, refetching } = getQuerySurface(reportQuery);
  const report = reportQuery.data;
  const blockedReason = downloadBlockedReason(report, reportQuery.isError, false);

  return (
    <DashboardProfileSection
      description="Payments taken and refunds made in each period: a refund counts on the day the money went back."
      id="manager-sales-periods"
      title="By period"
      variant="plain"
    >
      <div className="dashboard-page-stack">
        {report ? (
          <ManagerSalesChart report={report} />
        ) : reportQuery.isError ? null : (
          <div aria-hidden className="skeleton db-chart__skeleton" />
        )}
        <div className="dashboard-page-body">
          <div className="db-table-filters db-table-filters--end">
            <Button
              disabled={blockedReason !== undefined}
              size="sm"
              title={blockedReason}
              type="button"
              variant="outline"
              onClick={() => report && savePeriods(report)}
            >
              Download CSV
            </Button>
          </div>
          <DashboardTableWrap refetching={refetching}>
            <table className="db-table db-table--figures-4">
              <colgroup>
                <col />
                <col className="db-table-col-number" />
                <col className="db-table-col-number" />
                <col className="db-table-col-number" />
                <col className="db-table-col-number" />
              </colgroup>
              <thead>
                <tr>
                  <th>Period</th>
                  <th className="db-table-num">Orders</th>
                  <th className="db-table-num">Sales</th>
                  <th className="db-table-num">Refunds</th>
                  <th className="db-table-num">Net</th>
                </tr>
              </thead>
              <tbody>
                <DashboardTableStateRows
                  columnCount={COLUMN_COUNT}
                  entityLabel="periods"
                  isEmpty={false}
                  isError={reportQuery.isError}
                  initialLoading={initialLoading}
                />
                {report && !initialLoading && !reportQuery.isError
                  ? report.periods.map((period) => (
                      <tr key={period.start}>
                        <td className="db-table-cell-primary">{formatPeriod(period.start, report.group)}</td>
                        <td className="db-table-num">{period.orders}</td>
                        <td className="db-table-num">{formatPriceCents(period.gross_cents)}</td>
                        <td className="db-table-num">{formatPriceCents(period.refunds_cents)}</td>
                        <td className="db-table-num">{formatPriceCents(period.net_cents)}</td>
                      </tr>
                    ))
                  : null}
              </tbody>
            </table>
          </DashboardTableWrap>
        </div>
      </div>
    </DashboardProfileSection>
  );
}
