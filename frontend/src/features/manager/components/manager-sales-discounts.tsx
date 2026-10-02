"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { formatPriceCents } from "@/lib/validation/catalog";

import { useSalesReport } from "../hooks";
import type { SalesReportFilterInput } from "../schemas";

const COLUMN_COUNT = 3;

function orders(count: number): string {
  return count === 1 ? "1 order" : `${count} orders`;
}

type ManagerSalesDiscountsProps = {
  range: SalesReportFilterInput;
};

/** How many of the orders kept used a discount code, what the codes took off, and each code's part. */
export function ManagerSalesDiscounts({ range }: ManagerSalesDiscountsProps) {
  const reportQuery = useSalesReport(range);
  const { initialLoading, refetching } = getQuerySurface(reportQuery);
  const discounts = reportQuery.data?.discounts;
  const codes = discounts?.codes ?? [];

  return (
    <DashboardProfileSection
      description={
        discounts
          ? `${discounts.discounted_orders} of ${orders(discounts.orders)} paid and not cancelled used a code, ${formatPriceCents(discounts.discount_cents)} off in all.${codes.length > 0 ? " The ten codes that took off most:" : ""}`
          : "The orders that used a discount code, and the ten codes that took off most."
      }
      id="manager-sales-discounts"
      title="Discounts"
      variant="plain"
    >
      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--figures-2">
          <colgroup>
            <col />
            <col className="db-table-col-number" />
            <col className="db-table-col-number" />
          </colgroup>
          <thead>
            <tr>
              <th>Code</th>
              <th className="db-table-num">Orders</th>
              <th className="db-table-num">Discount</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={COLUMN_COUNT}
              emptyMessage="No discount code was used in these days."
              entityLabel="discount codes"
              isEmpty={codes.length === 0}
              isError={reportQuery.isError}
              initialLoading={initialLoading}
            />
            {!initialLoading && !reportQuery.isError
              ? codes.map((code) => (
                  <tr key={code.code}>
                    <td>
                      <span className="text-order-code">{code.code}</span>
                    </td>
                    <td className="db-table-num">{code.orders}</td>
                    <td className="db-table-num">{formatPriceCents(code.discount_cents)}</td>
                  </tr>
                ))
              : null}
          </tbody>
        </table>
      </DashboardTableWrap>
    </DashboardProfileSection>
  );
}
