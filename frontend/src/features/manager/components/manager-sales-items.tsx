"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { Button } from "@/components/ui/button";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { saveFile } from "@/lib/dom/save-file";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useSalesReport } from "../hooks";
import { centsToAmount, downloadBlockedReason, toCsv } from "../lib/sales-report";
import type { SalesReport, SalesReportFilterInput } from "../schemas";

const COLUMN_COUNT = 3;

const KIND_LABEL: Record<SalesReport["items"][number]["kind"], string> = {
  product: "Product",
  combo: "Combo",
};

/** The best sellers as a spreadsheet. */
function saveItems(report: SalesReport): void {
  const rows = [
    ["Item", "Kind", "Quantity", "Sales (USD)"],
    ...report.items.map((item) => [item.name, KIND_LABEL[item.kind], item.quantity, centsToAmount(item.sales_cents)]),
  ];
  saveFile(`best-sellers-${report.from}-to-${report.to}.csv`, toCsv(rows), "text/csv");
}

type ManagerSalesItemsProps = {
  range: SalesReportFilterInput;
};

/** The ten products and combos that sold most on the orders kept. */
export function ManagerSalesItems({ range }: ManagerSalesItemsProps) {
  const reportQuery = useSalesReport(range);
  const { initialLoading, refetching } = getQuerySurface(reportQuery);
  const report = reportQuery.data;
  const items = report?.items ?? [];
  const blockedReason = downloadBlockedReason(report, reportQuery.isError, items.length === 0);

  return (
    <DashboardProfileSection
      description="The ten that sold most on the orders paid in these days and not cancelled, at their prices before any discount."
      id="manager-sales-items"
      title="Best sellers"
      variant="plain"
    >
      <div className="dashboard-page-body">
        <div className="db-table-filters db-table-filters--end">
          <Button
            disabled={blockedReason !== undefined}
            size="sm"
            title={blockedReason}
            type="button"
            variant="outline"
            onClick={() => report && saveItems(report)}
          >
            Download CSV
          </Button>
        </div>
        <DashboardTableWrap refetching={refetching}>
          <table className="db-table db-table--figures-2">
            <colgroup>
              <col />
              <col className="db-table-col-number" />
              <col className="db-table-col-number" />
            </colgroup>
            <thead>
              <tr>
                <th>Item</th>
                <th className="db-table-num">Quantity</th>
                <th className="db-table-num">Sales</th>
              </tr>
            </thead>
            <tbody>
              <DashboardTableStateRows
                columnCount={COLUMN_COUNT}
                emptyMessage="Nothing sold in these days."
                entityLabel="items"
                isEmpty={items.length === 0}
                isError={reportQuery.isError}
                initialLoading={initialLoading}
              />
              {!initialLoading && !reportQuery.isError
                ? items.map((item) => (
                    <tr key={`${item.kind}:${item.id}`}>
                      <td>
                        <div className="db-table-stacked-cell min-w-0">
                          <span className="db-table-cell-primary truncate">{item.name}</span>
                          <span className="text-caption-dashboard text-muted">{KIND_LABEL[item.kind]}</span>
                        </div>
                      </td>
                      <td className="db-table-num">{item.quantity}</td>
                      <td className="db-table-num">{formatPriceCents(item.sales_cents)}</td>
                    </tr>
                  ))
                : null}
            </tbody>
          </table>
        </DashboardTableWrap>
      </div>
    </DashboardProfileSection>
  );
}
