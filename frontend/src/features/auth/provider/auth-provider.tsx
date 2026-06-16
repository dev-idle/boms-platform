"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, type ReactNode } from "react";

import { endLocalSession, ensureRefreshScheduled } from "@/lib/auth";
import { isApiError } from "@/lib/errors";
import { useAuthStore, waitForAuthStoreHydration } from "@/stores/auth-store";

import { MustChangePasswordGate } from "../components";
import { restoreSessionFromCookie } from "../lib/restore-session";

import { SessionHintProvider } from "./session-hint";

type AuthBootstrapEffectProps = {
  hasRefreshCookie: boolean;
};

const REFRESH_RATE_LIMIT_RETRY_MS = 2_000;
const REFRESH_RATE_LIMIT_MAX_ATTEMPTS = 3;

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

async function restoreSessionWithRetry(
  queryClient: Parameters<typeof restoreSessionFromCookie>[0],
): Promise<void> {
  let lastError: unknown;

  for (let attempt = 0; attempt < REFRESH_RATE_LIMIT_MAX_ATTEMPTS; attempt++) {
    try {
      await restoreSessionFromCookie(queryClient);
      return;
    } catch (error) {
      lastError = error;
      const canRetry =
        isApiError(error) &&
        error.isRateLimited() &&
        attempt < REFRESH_RATE_LIMIT_MAX_ATTEMPTS - 1;
      if (!canRetry) {
        throw error;
      }
      await sleep(REFRESH_RATE_LIMIT_RETRY_MS * (attempt + 1));
    }
  }

  throw lastError;
}

function AuthBootstrapEffect({ hasRefreshCookie }: AuthBootstrapEffectProps) {
  const queryClient = useQueryClient();
  const setStatus = useAuthStore((state) => state.setStatus);
  const bootIdRef = useRef(0);
  const rateLimitRetryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(
    null,
  );

  useEffect(() => {
    ensureRefreshScheduled();
  }, []);

  useEffect(() => {
    const bootId = ++bootIdRef.current;

    function clearRateLimitRetryTimer(): void {
      if (rateLimitRetryTimerRef.current !== null) {
        clearTimeout(rateLimitRetryTimerRef.current);
        rateLimitRetryTimerRef.current = null;
      }
    }

    async function bootstrap(): Promise<void> {
      await waitForAuthStoreHydration();
      if (bootId !== bootIdRef.current) {
        return;
      }

      const state = useAuthStore.getState();

      if (!hasRefreshCookie) {
        if (state.userSummary !== null) {
          endLocalSession();
        } else if (state.status === "idle") {
          setStatus("unauthenticated");
        }
        return;
      }

      if (state.status !== "idle") {
        ensureRefreshScheduled();
        return;
      }

      try {
        await restoreSessionWithRetry(queryClient);
        clearRateLimitRetryTimer();
      } catch (error) {
        if (bootId !== bootIdRef.current) {
          return;
        }
        if (isApiError(error) && error.isRateLimited()) {
          clearRateLimitRetryTimer();
          rateLimitRetryTimerRef.current = setTimeout(() => {
            if (bootId !== bootIdRef.current) {
              return;
            }
            if (useAuthStore.getState().status !== "idle") {
              return;
            }
            void bootstrap();
          }, REFRESH_RATE_LIMIT_RETRY_MS * 2);
          return;
        }
        clearRateLimitRetryTimer();
        endLocalSession();
      }
    }

    void bootstrap();

    return () => {
      clearRateLimitRetryTimer();
    };
  }, [hasRefreshCookie, queryClient, setStatus]);

  return null;
}

type AuthProviderProps = {
  /** From server `cookies()` in AuthBootstrap — single read per request. */
  hasRefreshCookie: boolean;
  children: ReactNode;
};

export function AuthProvider({ hasRefreshCookie, children }: AuthProviderProps) {
  return (
    <SessionHintProvider hasRefreshCookie={hasRefreshCookie}>
      <AuthBootstrapEffect hasRefreshCookie={hasRefreshCookie} />
      <MustChangePasswordGate />
      {children}
    </SessionHintProvider>
  );
}
