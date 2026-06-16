import { DASHBOARD_HOME_LEAD } from "@/constants/dashboard-page-copy";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.dashboard);

export default function ManagerDashboardPage() {
  return (
    <DashboardPageHeader
      description={DASHBOARD_HOME_LEAD.manager}
      title={PAGE_TITLES.dashboard}
    />
  );
}
