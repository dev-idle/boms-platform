/** True only on first fetch with no cached/placeholder data yet. */
export function isInitialQueryLoad(isPending: boolean, data: unknown): boolean {
  return isPending && data === undefined;
}

/** Background refetch — keep previous content visible under the busy overlay. */
export function isQueryRefetching(
  isFetching: boolean,
  isPending: boolean,
  data: unknown,
): boolean {
  return isFetching && !isPending && data !== undefined;
}

type DashboardQuerySurfaceInput = {
  data: unknown;
  isFetching: boolean;
  isPending: boolean;
};

/** Maps TanStack Query flags to dashboard async panel states. */
export function getDashboardQuerySurface({
  data,
  isFetching,
  isPending,
}: DashboardQuerySurfaceInput) {
  return {
    initialLoading: isInitialQueryLoad(isPending, data),
    refetching: isQueryRefetching(isFetching, isPending, data),
  };
}
