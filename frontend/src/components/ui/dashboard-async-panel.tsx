import type { ReactNode } from "react";

import { LOADING_MESSAGE } from "@/constants/loading-copy";
import { cn } from "@/lib/utils";

import { DashboardBusyOverlay } from "./dashboard-busy-overlay";

export type DashboardAsyncPanelProps = {
  children: ReactNode;
  className?: string;
  initialLoading?: boolean;
  /** When false, initial load is handled inline (e.g. table body rows). Default: true. */
  overlayOnInitialLoad?: boolean;
  refetching: boolean;
};

/**
 * Dashboard async shell — dots-ring overlay for refetch (and initial load when enabled).
 * Tables: initial load in tbody row 2; refetch uses `DashboardAsyncPanel` overlay.
 */
export function DashboardAsyncPanel({
  children,
  className,
  initialLoading = false,
  overlayOnInitialLoad = true,
  refetching,
}: DashboardAsyncPanelProps) {
  const overlayInitial = initialLoading && overlayOnInitialLoad;
  const showOverlay = refetching || overlayInitial;
  const busy = initialLoading || refetching;

  return (
    <div
      aria-busy={busy || undefined}
      aria-label={busy ? LOADING_MESSAGE : undefined}
      className={cn(
        "dashboard-async-panel",
        overlayInitial && "is-initial-loading",
        refetching && "is-refetching",
        className,
      )}
    >
      {showOverlay ? <DashboardBusyOverlay /> : null}
      {children}
    </div>
  );
}
