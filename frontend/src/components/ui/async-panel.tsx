import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

import { BusyIndicator } from "./loading-state";

type AsyncPanelProps = {
  children: ReactNode;
  className?: string;
  initialLoading?: boolean;
  /** When false, the caller draws the first load itself (e.g. a table's loading row). */
  overlayOnInitialLoad?: boolean;
  refetching: boolean;
};

/**
 * Content that loads and reloads in place, in either theme. A refetch dims the
 * content under a scrim carrying the busy indicator; the first load does the
 * same unless the caller draws its own. The scrim takes its ground from the
 * theme (`loading.css`), so one component serves dashboard and storefront.
 */
export function AsyncPanel({
  children,
  className,
  initialLoading = false,
  overlayOnInitialLoad = true,
  refetching,
}: AsyncPanelProps) {
  const overlayInitial = initialLoading && overlayOnInitialLoad;
  const showOverlay = refetching || overlayInitial;

  return (
    <div
      aria-busy={initialLoading || refetching || undefined}
      className={cn(
        "async-panel",
        overlayInitial && "is-initial-loading",
        refetching && "is-refetching",
        className,
      )}
    >
      {showOverlay ? (
        <div className="async-panel__overlay" role="status">
          <BusyIndicator />
        </div>
      ) : null}
      {children}
    </div>
  );
}
