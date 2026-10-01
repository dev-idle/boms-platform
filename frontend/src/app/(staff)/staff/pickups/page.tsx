import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { StaffPickupSchedule } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.pickups);

export default function StaffPickupsPage() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="The day's pickups by time. Hand a ready order over with the customer's code."
        eyebrow={DASHBOARD_PAGE_EYEBROW.operations}
        leadAside
        title={PAGE_TITLES.pickups}
      />
      <StaffPickupSchedule />
    </div>
  );
}
