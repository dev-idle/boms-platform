import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { StaffPrepQueue } from "@/features/staff";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.prepQueue);

export default function StaffPrepQueuePage() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="Counter tickets sorted by pickup time."
        eyebrow={DASHBOARD_PAGE_EYEBROW.operations}
        leadAside
        title={PAGE_TITLES.prepQueue}
      />
      <StaffPrepQueue />
    </div>
  );
}
