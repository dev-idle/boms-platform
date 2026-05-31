"use client";

import { refreshResponseSchema } from "@/features/auth/schemas";
import { ROUTE } from "@/constants/routes";
import { ApiError, ApiErrorCode } from "@/lib/errors";
import { useAuthStore } from "@/stores/auth-store";

import {
  clearStaleSession,
  isAuthSessionError,
  readApiEnvelope,
  throwApiErrorFromEnvelope,
} from "./session";

const REFRESH_BUFFER_MS = 60_000;

let refreshTimer: ReturnType<typeof setTimeout> | null = null;
let refreshPromise: Promise<void> | null = null;

export type RefreshOptions = {
  /** Redirect to login when refresh fails due to auth/session errors. */
  redirectOnFailure?: boolean;
};

export function scheduleRefresh(expiresAt: number | null): void {
  if (refreshTimer !== null) {
    clearTimeout(refreshTimer);
    refreshTimer = null;
  }
  if (expiresAt === null) {
    return;
  }
  const delay = expiresAt - Date.now() - REFRESH_BUFFER_MS;
  refreshTimer = setTimeout(() => {
    void refreshNow();
  }, Math.max(delay, 0));
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
  scheduleRefresh(useAuthStore.getState().expiresAt);
}

export function refreshNow(options: RefreshOptions = {}): Promise<void> {
  const { redirectOnFailure = true } = options;

  if (refreshPromise) {
    return refreshPromise;
  }

  refreshPromise = performRefresh()
    .catch(async (error) => {
      if (isAuthSessionError(error)) {
        await clearStaleSession();
      }
      useAuthStore.getState().clearAuth();
      if (refreshTimer !== null) {
        clearTimeout(refreshTimer);
        refreshTimer = null;
      }
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
  if (refreshTimer !== null) {
    clearTimeout(refreshTimer);
    refreshTimer = null;
  }
  refreshPromise = null;
}
