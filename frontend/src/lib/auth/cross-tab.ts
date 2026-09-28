import { z } from "zod";

/**
 * Every tab of the site shares one refresh cookie, and the API revokes the
 * whole session when a refresh token is spent twice. Tabs therefore take turns
 * to refresh, and hand the token a refresh produced to the others so a single
 * refresh serves them all.
 */
const REFRESH_LOCK = "boms:auth-refresh";
const TOKEN_CHANNEL = "boms:auth-token";

const sharedTokenSchema = z.object({
  access_token: z.string().min(1),
  expires_at: z.number().int().positive(),
});

export type SharedToken = z.infer<typeof sharedTokenSchema>;

export type TokenChannel = {
  share: (token: SharedToken) => void;
  close: () => void;
};

const accessTokenClaimsSchema = z.object({ sub: z.uuid() });

/**
 * The account an access token was issued to. Read without verification, only
 * to route a shared token to the tabs of that account: the API still verifies
 * every token it receives.
 */
export function accessTokenSubject(token: string): string | null {
  const payload = token.split(".")[1];
  if (!payload) {
    return null;
  }
  let claims: unknown;
  try {
    claims = JSON.parse(atob(payload.replace(/-/g, "+").replace(/_/g, "/")));
  } catch {
    return null;
  }
  const parsed = accessTokenClaimsSchema.safeParse(claims);
  return parsed.success ? parsed.data.sub : null;
}

/**
 * Runs task while no other tab of the site is refreshing. `waited` tells it
 * another tab held the lock first — and has probably just shared a token.
 */
export async function withRefreshLock(
  task: (waited: boolean) => Promise<void>,
): Promise<void> {
  if (typeof navigator === "undefined" || !("locks" in navigator)) {
    await task(false);
    return;
  }
  const ran = await navigator.locks.request(
    REFRESH_LOCK,
    { ifAvailable: true },
    async (lock) => {
      if (lock === null) {
        return false;
      }
      await task(false);
      return true;
    },
  );
  if (!ran) {
    await navigator.locks.request(REFRESH_LOCK, () => task(true));
  }
}

/** Connects this tab to the others: `share` reaches every other tab's `onToken`. */
export function createTokenChannel(
  onToken: (token: SharedToken) => void,
): TokenChannel | null {
  if (typeof BroadcastChannel === "undefined") {
    return null;
  }
  const channel = new BroadcastChannel(TOKEN_CHANNEL);
  channel.onmessage = (message: MessageEvent<unknown>) => {
    const parsed = sharedTokenSchema.safeParse(message.data);
    // Any other shape comes from a tab running another build of the site.
    if (parsed.success) {
      onToken(parsed.data);
    }
  };
  return {
    share: (token) => channel.postMessage(token),
    close: () => channel.close(),
  };
}
