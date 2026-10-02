import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { DASHBOARD_HOME_LEAD, DASHBOARD_PAGE_EYEBROW } from "@/constants/dashboard-page-copy";
import { PAGE_TITLES } from "@/lib/metadata/page-title";

import { ManagerOperationsKpis } from "./manager-operations-kpis";
import { ManagerOperationsLate } from "./manager-operations-late";
import { ManagerOperationsWork } from "./manager-operations-work";

/** The manager's home: the bakery's work as it moves — open orders, stations, today's pickups and the late ones. */
export function ManagerOperations() {
  return (
    <div className="dashboard-page-stack">
      <DashboardPageHeader
        description={DASHBOARD_HOME_LEAD.manager}
        eyebrow={DASHBOARD_PAGE_EYEBROW.overview}
        title={PAGE_TITLES.dashboard}
      />
      <ManagerOperationsKpis />
      {/* Block flow, so the sections keep their own rhythm and hairlines. */}
      <div>
        <ManagerOperationsLate />
        <ManagerOperationsWork />
      </div>
    </div>
  );
}
