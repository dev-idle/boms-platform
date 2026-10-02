import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { ENGAGEMENT_GOAL_PERCENT } from "../lib/engagement";
import { ManagerEngagementFeatures } from "./manager-engagement-features";
import { ManagerEngagementKpis } from "./manager-engagement-kpis";

/** How registered customers use reviews, favorites, the wishlist, messages and promotion emails. */
export function ManagerEngagement() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description={`How registered customers use reviews, favorites, the wishlist, messages and promotion emails. The goal is ${ENGAGEMENT_GOAL_PERCENT}% using two or more.`}
        eyebrow={DASHBOARD_PAGE_EYEBROW.reports}
        title={PAGE_TITLES.engagement}
      />
      <ManagerEngagementKpis />
      <ManagerEngagementFeatures />
    </div>
  );
}
