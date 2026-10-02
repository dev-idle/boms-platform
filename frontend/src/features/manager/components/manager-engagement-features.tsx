"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { getQuerySurface } from "@/lib/react-query/query-surface";

import { useEngagementReport } from "../hooks";
import { shareOf } from "../lib/share";
import type { EngagementFeature } from "../schemas";

const COLUMN_COUNT = 3;

const FEATURE_LABEL: Record<EngagementFeature, string> = {
  reviews: "Reviews",
  favorites: "Favorites",
  wishlist: "Wishlist",
  messages: "Messages",
  promotions: "Promotion emails",
};

/** How many registered customers used each feature, and their share. */
export function ManagerEngagementFeatures() {
  const reportQuery = useEngagementReport();
  const { initialLoading, refetching } = getQuerySurface(reportQuery);
  const report = reportQuery.data;

  return (
    <DashboardProfileSection
      description="A review or a message written, or a product saved to a list, at least once; promotion emails count the customers who agree to them now."
      id="manager-engagement-features"
      title="By feature"
      variant="plain"
    >
      <DashboardTableWrap refetching={refetching}>
        <table className="db-table">
          <colgroup>
            <col />
            <col className="db-table-col-number" />
            <col className="db-table-col-number" />
          </colgroup>
          <thead>
            <tr>
              <th>Feature</th>
              <th className="db-table-num">Customers</th>
              <th className="db-table-num">Share</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={COLUMN_COUNT}
              entityLabel="features"
              isEmpty={false}
              isError={reportQuery.isError}
              initialLoading={initialLoading}
            />
            {report && !initialLoading && !reportQuery.isError
              ? report.features.map(({ feature, customers }) => (
                  <tr key={feature}>
                    <td className="db-table-cell-primary">{FEATURE_LABEL[feature]}</td>
                    <td className="db-table-num">{customers}</td>
                    <td className="db-table-num">{shareOf(customers, report.customers)}</td>
                  </tr>
                ))
              : null}
          </tbody>
        </table>
      </DashboardTableWrap>
    </DashboardProfileSection>
  );
}
