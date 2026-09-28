"use client";

import { refreshResponseSchema } from "@/lib/schemas/auth";
import { ROUTE } from "@/constants/routes";
import {
  ApiError,
  ApiErrorCode,
  isApiError,
  throwApiErrorFromEnvelope,
} from "@/lib/errors";
import { useAuthStore } from "@/stores/auth-store";

import {
  accessTokenSubject,
  createTokenChannel,
  withRefreshLock,
  type SharedToken,
  type TokenChannel,
} from "./cross-tab";
import {
  clearStaleSession,
  isAuthSessionError,
  readApiEnvelope,
} from "./session";

const REFRESH_BUFFER_MS = 60_000;
/**
 * How long a tab that queued behind another tab's refresh waits for the token
 * that refresh shares before refreshing itself. The lock and the channel are
 * separate, so the token may land a moment after the lock does.
 */
const SHARED_TOKEN_WAIT_MS = 1_000;

let refreshTimer: ReturnType<typeof setTimeout> | null = null;
let refreshPromise: Promise<void> | null = null;
let tokenChannel: TokenChannel | null = null;
const tokenWatchers = new Set<() => void>();

export type RefreshOptions = {
  /** Redirect to login when refresh fails due to auth/session errors. */
  redirectOnFailure?: boolean;
  /**
   * The access token that was refused or is about to expire; defaults to the
   * tab's current one. When the tab already holds a different token by the
   * time its turn comes, another tab refreshed and the refresh is skipped.
   */
  staleToken?: string | null;
};

/** Takes a token another tab refreshed, if it belongs to this tab's account. */
function adoptSharedToken(token: SharedToken): void {
  const { status, user, setTokens } = useAuthStore.getState();
  // A sign-in elsewhere replaces the cookie; that account's token is not ours.
  if (
    status !== "authenticated" ||
    user === null ||
    accessTokenSubject(token.access_token) !== user.id
  ) {
    return;
  }
  // A late or reordered message must not replace a newer token.
  if (token.expires_at <= (useAuthStore.getState().expiresAt ?? 0)) {
    return;
  }
  setTokens({
    accessToken: token.access_token,
    expiresIn: (token.expires_at - Date.now()) / 1000,
  });
  scheduleRefresh(useAuthStore.getState().expiresAt);
  for (const notify of tokenWatchers) {
    notify();
  }
}

/** Resolves true once the tab holds a token other than stale, false after ms. */
function tokenReplaced(stale: string | null, ms: number): Promise<boolean> {
  return new Promise((resolve) => {
    const finish = (replaced: boolean) => {
      clearTimeout(timer);
      tokenWatchers.delete(check);
      resolve(replaced);
    };
    const check = () => {
      if (useAuthStore.getState().accessToken !== stale) {
        finish(true);
      }
    };
    const timer = setTimeout(() => finish(false), ms);
    tokenWatchers.add(check);
    check();
  });
}

function sharedTokens(): TokenChannel | null {
  tokenChannel ??= createTokenChannel(adoptSharedToken);
  return tokenChannel;
}

export function scheduleRefresh(expiresAt: number | null): void {
  sharedTokens();
  if (refreshTimer !== null) {
    clearTimeout(refreshTimer);
    refreshTimer = null;
  }
  if (expiresAt === null) {
    return;
  }
  const delay = expiresAt - Date.now() - REFRESH_BUFFER_MS;
  const armedFor = useAuthStore.getState().accessToken;
  refreshTimer = setTimeout(() => {
    void refreshNow({ staleToken: armedFor });
  }, Math.max(delay, 0));
}

/** Clears Zustand auth state and persisted sessionStorage hint (no timers, no API). */
export function clearLocalAuthState(): void {
  useAuthStore.getState().clearAuth();
  void useAuthStore.persist.clearStorage();
}

/** Re-arm proactive refresh after remount (e.g. React Strict Mode). */
export function ensureRefreshScheduled(): void {
  const { status, expiresAt } = useAuthStore.getState();
  if (status === "authenticated" && expiresAt !== null) {
    scheduleRefresh(expiresAt);
  }
}

async function performRefresh(): Promise<void> {
  const response = await fetch("/api/v1/auth/refresh", {
    method: "POST",
    credentials: "include",
    headers: { Accept: "application/json" },
    cache: "no-store",
  });

  const { status, envelope } = await readApiEnvelope(response);

  if (!response.ok || !envelope.success) {
    throwApiErrorFromEnvelope(status, envelope);
  }

  const parsed = refreshResponseSchema.safeParse(envelope.data);
  if (!parsed.success) {
    throw new ApiError(502, {
      code: ApiErrorCode.InvalidResponse,
      message: "Refresh response failed schema validation",
    });
  }

  const { setTokens, user, updateUser } = useAuthStore.getState();
  setTokens({
    accessToken: parsed.data.access_token,
    expiresIn: parsed.data.expires_in,
  });
  if (user && parsed.data.must_change_password !== undefined) {
    updateUser({ ...user, must_change_password: parsed.data.must_change_password });
  }
  const { expiresAt } = useAuthStore.getState();
  scheduleRefresh(expiresAt);
  // The refresh retired the session the other tabs' tokens name.
  if (expiresAt !== null) {
    sharedTokens()?.share({
      access_token: parsed.data.access_token,
      expires_at: expiresAt,
    });
  }
}

function isTransientRefreshError(error: unknown): boolean {
  return isApiError(error) && error.isRateLimited();
}

export function refreshNow(options: RefreshOptions = {}): Promise<void> {
  const { redirectOnFailure = true } = options;
  const stale =
    options.staleToken === undefined
      ? useAuthStore.getState().accessToken
      : options.staleToken;

  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = withRefreshLock(async (waited) => {
    // Another tab refreshed first and its token is here, or is on its way.
    if (stale !== null) {
      const wait = waited ? SHARED_TOKEN_WAIT_MS : 0;
      if (await tokenReplaced(stale, wait)) {
        return;
      }
    }
    await performRefresh();
  })
    .catch(async (error) => {
      if (isTransientRefreshError(error)) {
        throw error;
      }

      if (
        isAuthSessionError(error) &&
        !(
          isApiError(error) &&
          error.code === ApiErrorCode.MissingRefreshToken
        )
      ) {
        await clearStaleSession();
      }
      resetRefreshManager();
      clearLocalAuthState();
      if (
        redirectOnFailure &&
        isAuthSessionError(error) &&
        typeof window !== "undefined"
      ) {
        window.location.href = ROUTE.login;
      }
      throw error;
    })
    .finally(() => {
      refreshPromise = null;
    });

  return refreshPromise;
}

/** Clears refresh timer and in-flight refresh state (logout). */
export function resetRefreshManager(): void {
  // A signed-out tab has no use for other tabs' tokens.
  tokenChannel?.close();
  tokenChannel = null;
  if (refreshTimer !== null) {
    clearTimeout(refreshTimer);
    refreshTimer = null;
  }
  refreshPromise = null;
}
