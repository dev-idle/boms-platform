import { getMe } from "@/features/user";
import { refreshNow, scheduleRefresh } from "@/lib/auth";
import { useAuthStore } from "@/stores/auth-store";

/**
 * Exchanges HttpOnly refresh cookie for access token + loads /me.
 * Caller must ensure refresh cookie is present (server hint) before calling.
 */
export async function restoreSessionFromCookie(): Promise<void> {
  await refreshNow({ redirectOnFailure: false });

  const user = await getMe();
  const { accessToken, expiresAt, setAuth } = useAuthStore.getState();

  if (!accessToken || expiresAt === null) {
    throw new Error("Missing session after refresh");
  }

  setAuth({
    accessToken,
    expiresIn: Math.max(Math.floor((expiresAt - Date.now()) / 1000), 1),
    user,
  });
  scheduleRefresh(useAuthStore.getState().expiresAt);
}
