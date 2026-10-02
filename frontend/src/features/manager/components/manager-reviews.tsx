import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { ManagerProductRatings } from "./manager-product-ratings";
import { ManagerReviewKpis } from "./manager-review-kpis";
import { ManagerReviewsTable } from "./manager-reviews-table";

/** Customers' reviews of what they picked up: the numbers, the reviews to moderate, and each product's rating. */
export function ManagerReviews() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="Publish what customers wrote about their pickups, hide what breaks the rules, and see how each product is rated."
        eyebrow={DASHBOARD_PAGE_EYEBROW.feedback}
        title={PAGE_TITLES.reviews}
      />
      <ManagerReviewKpis />
      <div className="dashboard-page-body">
        <ManagerReviewsTable />
      </div>
      <ManagerProductRatings />
    </div>
  );
}
