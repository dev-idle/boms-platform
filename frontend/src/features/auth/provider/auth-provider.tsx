"use client";

import { useEffect, useRef, type ReactNode } from "react";

import { endLocalSession, ensureRefreshScheduled } from "@/lib/auth";
import { useAuthStore, waitForAuthStoreHydration } from "@/stores/auth-store";

import { MustChangePasswordGate } from "../components";
import { restoreSessionFromCookie } from "../lib/restore-session";

import { SessionHintProvider } from "./session-hint";

type AuthBootstrapEffectProps = {
  hasRefreshCookie: boolean;
};

function AuthBootstrapEffect({ hasRefreshCookie }: AuthBootstrapEffectProps) {
  const setStatus = useAuthStore((state) => state.setStatus);
  const bootIdRef = useRef(0);

  useEffect(() => {
    ensureRefreshScheduled();
  }, []);

  useEffect(() => {
    const bootId = ++bootIdRef.current;

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
        await restoreSessionFromCookie();
      } catch {
        if (bootId === bootIdRef.current) {
          endLocalSession();
        }
      }
    }

    void bootstrap();
  }, [hasRefreshCookie, setStatus]);

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
