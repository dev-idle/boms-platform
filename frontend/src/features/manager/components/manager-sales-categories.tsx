"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useSalesReport } from "../hooks";
import { shareOf } from "../lib/share";
import type { SalesReportFilterInput } from "../schemas";

const COLUMN_COUNT = 4;

type ManagerSalesCategoriesProps = {
  range: SalesReportFilterInput;
};

/** What each category sold on the orders kept, and its share of it all. */
export function ManagerSalesCategories({ range }: ManagerSalesCategoriesProps) {
  const reportQuery = useSalesReport(range);
  const { initialLoading, refetching } = getQuerySurface(reportQuery);
  const categories = reportQuery.data?.categories ?? [];
  const total = categories.reduce((sum, category) => sum + category.sales_cents, 0);

  return (
    <DashboardProfileSection
      description="Each product counts in the category it is in now; combos count on their own."
      id="manager-sales-categories"
      title="By category"
      variant="plain"
    >
      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--figures-3">
          <colgroup>
            <col />
            <col className="db-table-col-number" />
            <col className="db-table-col-number" />
            <col className="db-table-col-number" />
          </colgroup>
          <thead>
            <tr>
              <th>Category</th>
              <th className="db-table-num">Quantity</th>
              <th className="db-table-num">Sales</th>
              <th className="db-table-num">Share</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={COLUMN_COUNT}
              emptyMessage="Nothing sold in these days."
              entityLabel="categories"
              isEmpty={categories.length === 0}
              isError={reportQuery.isError}
              initialLoading={initialLoading}
            />
            {!initialLoading && !reportQuery.isError
              ? categories.map((category) => (
                  <tr key={category.category_id ?? "combos"}>
                    <td className="db-table-cell-primary">{category.name ?? "Combos"}</td>
                    <td className="db-table-num">{category.quantity}</td>
                    <td className="db-table-num">{formatPriceCents(category.sales_cents)}</td>
                    <td className="db-table-num">{shareOf(category.sales_cents, total)}</td>
                  </tr>
                ))
              : null}
          </tbody>
        </table>
      </DashboardTableWrap>
    </DashboardProfileSection>
  );
}
