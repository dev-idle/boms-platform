/** True only on first fetch with no cached/placeholder data yet. */
export function isInitialQueryLoad(isPending: boolean, data: unknown): boolean {
  return isPending && data === undefined;
}

/** Background refetch — keep previous rows visible, dim table slightly. */
export function isQueryRefetching(
  isFetching: boolean,
  isPending: boolean,
  data: unknown,
): boolean {
  return isFetching && !isPending && data !== undefined;
}
