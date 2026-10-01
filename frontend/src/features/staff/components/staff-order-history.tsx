"use client";

import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";
import {
  formatOrderStatusLabel,
  orderStatusToPillVariant,
  StatusPill,
} from "@/components/ui/status-pill";
import { roleDisplayLabel } from "@/constants/roles";
import { formatDateTime } from "@/lib/validation/datetime";

import type { StaffOrder } from "../schemas";

type StaffOrderHistoryProps = {
  timeline: StaffOrder["timeline"];
};

/** Every status the order entered, when, and which role — or the system — moved it there. */
export function StaffOrderHistory({ timeline }: StaffOrderHistoryProps) {
  return (
    <DashboardProfileSection id="staff-order-history" title="Status history" variant="plain">
      <div className="dashboard-activity-feed-panel">
        <ol className="dashboard-activity-feed">
          {timeline.map((entry) => (
            <li key={`${entry.status}-${entry.at}`} className="dashboard-activity-feed-item">
              <div className="dashboard-activity-feed-head">
                <StatusPill
                  label={formatOrderStatusLabel(entry.status)}
                  variant={orderStatusToPillVariant(entry.status)}
                />
                <time className="dashboard-activity-feed-when" dateTime={entry.at}>
                  {formatDateTime(entry.at)}
                </time>
              </div>
              <p className="dashboard-activity-feed-meta">
                {entry.actor_role ? roleDisplayLabel(entry.actor_role) : "System"}
              </p>
              {entry.reason ? <p className="dashboard-activity-feed-meta">Reason: {entry.reason}</p> : null}
            </li>
          ))}
        </ol>
      </div>
    </DashboardProfileSection>
  );
}
