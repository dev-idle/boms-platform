"use client";

import { Button } from "@/components/ui/button";

import { useAdminDashboard } from "../hooks";
import type { AdminDashboard } from "../schemas";

const KPIS: ReadonlyArray<{ label: string; value: (figures: AdminDashboard) => number }> = [
  { label: "Accounts in use", value: ({ accounts }) => accounts.by_role.reduce((sum, role) => sum + role.count, 0) },
  { label: "New customers this week", value: ({ accounts }) => accounts.new_customers_this_week },
  { label: "Unconfirmed emails", value: ({ accounts }) => accounts.unconfirmed_emails },
  { label: "Disabled accounts", value: ({ accounts }) => accounts.disabled },
];

/** The accounts in four numbers: in use, new this week, unconfirmed, disabled. */
export function AdminDashboardKpis() {
  const dashboardQuery = useAdminDashboard();
  const figures = dashboardQuery.data;

  if (dashboardQuery.isError) {
    return (
      <div className="db-kpi-error">
        <p className="db-kpi-error__title">We could not read the accounts and the system&apos;s health.</p>
        <Button type="button" variant="outline" onClick={() => void dashboardQuery.refetch()}>
          Try again
        </Button>
      </div>
    );
  }

  return (
    <dl aria-busy={figures ? undefined : true} className="db-kpi-strip">
      {KPIS.map((kpi) => (
        <div className="db-kpi" key={kpi.label}>
          <dt className="db-kpi__label">{kpi.label}</dt>
          <dd className="db-kpi__value">
            {figures ? kpi.value(figures) : <span className="skeleton db-kpi__skeleton" />}
          </dd>
        </div>
      ))}
    </dl>
  );
}
