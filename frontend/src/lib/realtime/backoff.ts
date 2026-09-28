const BASE_DELAY_MS = 1_000;
const MAX_DELAY_MS = 30_000;

/**
 * Delay before reconnect attempt `attempt` (0-based): exponential up to 30 s,
 * with half of it randomised so tabs that lost the server together do not all
 * come back in the same instant.
 */
export function reconnectDelay(
  attempt: number,
  random: () => number = Math.random,
): number {
  const ceiling = Math.min(MAX_DELAY_MS, BASE_DELAY_MS * 2 ** attempt);
  return Math.round(ceiling / 2 + random() * (ceiling / 2));
}
