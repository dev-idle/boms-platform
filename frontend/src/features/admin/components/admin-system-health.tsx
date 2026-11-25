"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import { DashboardTableStateRows } from "@/components/ui/dashboard-table-state-rows";
import { DashboardTableWrap } from "@/components/ui/dashboard-table-wrap";
import { StatusPill, type StatusPillVariant } from "@/components/ui/status-pill";
import { getQuerySurface } from "@/lib/react-query/query-surface";
import { plural } from "@/lib/utils";
import { formatDateTime } from "@/lib/validation/datetime";

import { useAdminDashboard } from "../hooks";
import type { AdminDashboard } from "../schemas";

type Health = AdminDashboard["health"];

type Check = { name: string; state: { label: string; variant: StatusPillVariant }; detail: string };

const UP = { label: "Up", variant: "ready" } as const;
const DOWN = { label: "Down", variant: "cancelled" } as const;
const CLEAR = { label: "Clear", variant: "ready" } as const;
const ATTENTION = { label: "Needs attention", variant: "cancelled" } as const;
const UNKNOWN = { label: "Unknown", variant: "completed" } as const;

/** Each check in the order the table lists them: what runs, then what has piled up. */
function checks(health: Health): Check[] {
  return [
    { name: "Database", state: health.database === "up" ? UP : DOWN, detail: "The bakery's orders, accounts and catalog." },
    {
      name: "Redis",
      state: health.redis === "up" ? UP : DOWN,
      detail: "Sessions, sign-in limits, live updates and the email queue.",
    },
    {
      name: "Worker",
      state:
        health.worker === "up"
          ? { label: "Running", variant: "ready" }
          : health.worker === "late"
            ? { label: "Late", variant: "pending" }
            : UNKNOWN,
      detail: health.worker_last_run_at
        ? `Last ran ${formatDateTime(health.worker_last_run_at)}.`
        : "No run on record: it may be stopped, or Redis cannot be read.",
    },
    {
      name: "Undelivered events",
      state: health.undelivered_events === 0 ? CLEAR : ATTENTION,
      detail: `${plural(health.undelivered_events, "event")} waiting more than five minutes to reach staff, customers or email.`,
    },
    {
      name: "Failed emails",
      state: health.failed_emails === null ? UNKNOWN : health.failed_emails === 0 ? CLEAR : ATTENTION,
      detail:
        health.failed_emails === null
          ? "The email queue cannot be read while Redis is down."
          : `${plural(health.failed_emails, "email")} given up on after the last try, kept for up to 90 days.`,
    },
    {
      name: "Stuck refunds",
      state: health.stuck_refunds === 0 ? CLEAR : ATTENTION,
      detail: `${plural(health.stuck_refunds, "refund")} asked for more than an hour ago and not made yet.`,
    },
  ];
}

/** Whether the system runs as it should: what it runs on, and what has piled up. */
export function AdminSystemHealth() {
  const dashboardQuery = useAdminDashboard();
  const { initialLoading, refetching } = getQuerySurface(dashboardQuery);
  // A failed refresh keeps the last figures; they are not shown as current.
  const health = dashboardQuery.isError ? undefined : dashboardQuery.data?.health;

  return (
    <DashboardProfileSection
      description="What the system runs on, then what has piled up."
      id="admin-system-health"
      title="System health"
      variant="plain"
    >
      <DashboardTableWrap refetching={refetching}>
        <table className="db-table db-table--health">
          <colgroup>
            <col className="db-table-col-name" />
            <col className="db-table-col-status" />
            <col />
          </colgroup>
          <thead>
            <tr>
              <th>Check</th>
              <th>Status</th>
              <th>Detail</th>
            </tr>
          </thead>
          <tbody>
            <DashboardTableStateRows
              columnCount={3}
              entityLabel="checks"
              initialLoading={initialLoading}
              isEmpty={false}
              isError={dashboardQuery.isError}
            />
            {health
              ? checks(health).map((check) => (
                  <tr key={check.name}>
                    <td className="db-table-cell-primary">{check.name}</td>
                    <td>
                      <StatusPill label={check.state.label} variant={check.state.variant} />
                    </td>
                    <td className="text-muted">{check.detail}</td>
                  </tr>
                ))
              : null}
          </tbody>
        </table>
      </DashboardTableWrap>
    </DashboardProfileSection>
  );
}
