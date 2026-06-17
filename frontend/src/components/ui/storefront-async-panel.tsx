import type { ReactNode } from "react";

import { LOADING_MESSAGE } from "@/constants/loading-copy";
import { cn } from "@/lib/utils";

import { StorefrontBusyOverlay } from "./storefront-busy-overlay";

export type StorefrontAsyncPanelProps = {
  children: ReactNode;
  className?: string;
  initialLoading?: boolean;
  /** When false, initial load is handled inline (e.g. centered spinner). Default: true. */
  overlayOnInitialLoad?: boolean;
  refetching: boolean;
};

/** Storefront async shell — dots-ring overlay on refetch (dashboard parity). */
export function StorefrontAsyncPanel({
  children,
  className,
  initialLoading = false,
  overlayOnInitialLoad = true,
  refetching,
}: StorefrontAsyncPanelProps) {
  const overlayInitial = initialLoading && overlayOnInitialLoad;
  const showOverlay = refetching || overlayInitial;
  const busy = initialLoading || refetching;

  return (
    <div
      aria-busy={busy || undefined}
      aria-label={busy ? LOADING_MESSAGE : undefined}
      className={cn(
        "storefront-async-panel",
        overlayInitial && "is-initial-loading",
        refetching && "is-refetching",
        className,
      )}
    >
      {showOverlay ? <StorefrontBusyOverlay /> : null}
      {children}
    </div>
  );
}
