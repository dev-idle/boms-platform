import { InlineLoadingState, PageLoadingState } from "@/components/ui/loading-state";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";

/**
 * Root app fallback — needs a theme when no route-group layout is mounted yet.
 * Page variant, the same as the root layout's Suspense fallback, so the two
 * never swap the spinner between two positions on first load.
 */
export function RootRouteLoading() {
  return (
    <ThemeScope theme={APP_THEME.storefront}>
      <PageLoadingState />
    </ThemeScope>
  );
}

/** Inside a route-group layout — inherits `data-theme` from parent; no extra wrapper. */
export function SegmentRouteLoading() {
  return <InlineLoadingState />;
}
