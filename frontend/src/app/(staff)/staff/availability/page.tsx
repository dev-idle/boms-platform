import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { StaffAvailability } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.availability);

export default function StaffAvailabilityPage() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="Mark what has run out today. It can still be ordered for a later day."
        eyebrow={DASHBOARD_PAGE_EYEBROW.operations}
        leadAside
        title={PAGE_TITLES.availability}
      />
      <StaffAvailability />
    </div>
  );
}
