import { LoadingState } from "@/components/ui/loading-state";
import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";

/** Root app fallback — needs theme when no route-group layout is mounted yet. */
export function RootRouteLoading() {
  return (
    <ThemeScope theme={APP_THEME.storefront}>
      <SegmentRouteLoading />
    </ThemeScope>
  );
}

/** Inside a route-group layout — inherits `data-theme` from parent; no extra wrapper. */
export function SegmentRouteLoading() {
  return <LoadingState showMessage={false} variant="inline" />;
}
