import {
  DASHBOARD_HOME_LEAD,
  DASHBOARD_PAGE_EYEBROW,
} from "@/constants/dashboard-page-copy";
import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { PAGE_TITLES, pageTitle } from "@/lib/metadata/page-title";

export const metadata = pageTitle(PAGE_TITLES.dashboard);

export default function AdminDashboardPage() {
  return (
    <DashboardPageHeader
      description={DASHBOARD_HOME_LEAD.admin}
      eyebrow={DASHBOARD_PAGE_EYEBROW.overview}
      title={PAGE_TITLES.dashboard}
    />
  );
}
