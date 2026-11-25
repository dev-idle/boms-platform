"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { roleDisplayLabel } from "@/constants/roles";
import { getQuerySurface } from "@/lib/react-query/query-surface";

import { useAdminDashboard } from "../hooks";

/** The accounts in use by role, every role listed. */
export function AdminAccounts() {
  const dashboardQuery = useAdminDashboard();
  const { initialLoading, refetching } = getQuerySurface(dashboardQuery);
  // A failed refresh keeps the last figures; they are not shown as current.
  const loaded = dashboardQuery.data !== undefined && !dashboardQuery.isError;

  return (
    <DashboardProfileSection
      description="Disabled accounts are not counted under any role."
      id="admin-accounts"
      title="Accounts"
      variant="plain"
    >
      <DashboardTableWrap refetching={refetching}>
        <table className="db-table">
          <colgroup>
            <col />
            <col className="db-table-col-number" />
          </colgroup>
          <thead>
            <tr>
              <th>Role</th>
              <th className="db-table-num">Accounts</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={2}
              entityLabel="accounts"
              initialLoading={initialLoading}
              isEmpty={false}
              isError={dashboardQuery.isError}
            />
            {loaded
              ? dashboardQuery.data.accounts.by_role.map((row) => (
                  <tr key={row.role}>
                    <td className="db-table-cell-primary">{roleDisplayLabel(row.role)}</td>
                    <td className="db-table-num">{row.count}</td>
                  </tr>
                ))
              : null}
          </tbody>
        </table>
      </DashboardTableWrap>
    </DashboardProfileSection>
  );
}
