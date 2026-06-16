import { LOADING_MESSAGE } from "@/constants/loading-copy";

import { LoadingIndicator } from "./loading-state";

/** Canonical dashboard busy affordance — dots ring, dense tables and feeds. */
export const DASHBOARD_BUSY_DOTS = 6;

export function DashboardBusyIndicator() {
  return (
    <>
      <span className="sr-only">{LOADING_MESSAGE}</span>
      <LoadingIndicator dots={DASHBOARD_BUSY_DOTS} />
    </>
  );
}

export function DashboardBusyOverlay() {
  return (
    <div className="dashboard-async-panel__overlay" role="status">
      <DashboardBusyIndicator />
    </div>
  );
}
