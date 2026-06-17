import { LOADING_MESSAGE } from "@/constants/loading-copy";

import { LoadingIndicator } from "./loading-state";

const STOREFRONT_BUSY_DOTS = 6;

export function StorefrontBusyIndicator() {
  return (
    <>
      <span className="sr-only">{LOADING_MESSAGE}</span>
      <LoadingIndicator dots={STOREFRONT_BUSY_DOTS} />
    </>
  );
}

export function StorefrontBusyOverlay() {
  return (
    <div className="storefront-async-panel__overlay" role="status">
      <StorefrontBusyIndicator />
    </div>
  );
}
