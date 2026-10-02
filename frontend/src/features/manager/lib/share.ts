/**
 * `count` as a whole percentage of `total`, rounded down so a share never
 * reads as reaching a goal it has not; "—" when there is nothing to share.
 */
export function shareOf(count: number, total: number): string {
  return total === 0 ? "—" : `${Math.floor((count / total) * 100)}%`;
}
