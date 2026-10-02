/**
 * The goal objective 4.3.1 sets: the share of registered customers using two
 * or more engagement features.
 */
export const ENGAGEMENT_GOAL_PERCENT = 60;

/**
 * `count` as a whole percentage of `total`, rounded down so a share never
 * reads as reaching a goal it has not; "—" when there is nobody to count.
 */
export function shareOf(count: number, total: number): string {
  return total === 0 ? "—" : `${Math.floor((count / total) * 100)}%`;
}
