import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { BakerProductionTable } from "@/features/baker";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.production);

export default function BakerProductionPage() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description="Kitchen queue sorted by pickup time."
        title={PAGE_TITLES.production}
      />
      <BakerProductionTable />
    </div>
  );
}
