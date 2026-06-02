import { clearLocalAuthState, resetRefreshManager } from "./refresh-manager";

/** Clears in-memory tokens, persisted client hints, and refresh timers (no API). */
export function endLocalSession(): void {
  resetRefreshManager();
  clearLocalAuthState();
}
