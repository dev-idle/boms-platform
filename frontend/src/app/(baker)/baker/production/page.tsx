import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { BakerProductionTable } from "@/features/baker";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.production);

export default function BakerProductionPage() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="The kitchen's tickets, the soonest pickup first."
        eyebrow={DASHBOARD_PAGE_EYEBROW.operations}
        title={PAGE_TITLES.production}
      />
      <BakerProductionTable />
    </div>
  );
}
