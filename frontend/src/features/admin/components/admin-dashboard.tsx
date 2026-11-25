import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_HOME_LEAD, DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { AdminAccounts } from "./admin-accounts";
import { AdminDashboardKpis } from "./admin-dashboard-kpis";
import { AdminSystemHealth } from "./admin-system-health";

/** The admin's home: the accounts in use, and whether the system runs as it should. */
export function AdminDashboard() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description={DASHBOARD_HOME_LEAD.admin}
        eyebrow={DASHBOARD_PAGE_EYEBROW.overview}
        title={PAGE_TITLES.dashboard}
      />
      <AdminDashboardKpis />
      {/* Block flow, so the sections keep their own rhythm and hairlines. */}
      <div>
        <AdminAccounts />
        <AdminSystemHealth />
      </div>
    </div>
  );
}
