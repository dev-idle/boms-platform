type QuerySurfaceInput = {
  isPending: boolean;
  isPlaceholderData: boolean;
};

/**
 * Maps TanStack Query flags to what an `AsyncPanel` shows.
 *
 * `initialLoading` — nothing to show yet. `refetching` — what is shown is the
 * previous page or filter's data, kept as a placeholder while the new one
 * loads, so it is dimmed and locked. A background refresh of data that is
 * already right (a remount past `staleTime`, an invalidation after a save)
 * stays silent: the rows are correct and simply update in place.
 */
export function getQuerySurface({ isPending, isPlaceholderData }: QuerySurfaceInput) {
  return {
    initialLoading: isPending,
    refetching: isPlaceholderData,
  };
}
