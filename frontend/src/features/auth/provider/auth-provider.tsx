"use client";

import { useEffect, type ReactNode } from "react";

import { useAuthStore } from "@/stores/auth-store";

import { getMe } from "../api";
import {
  ensureRefreshScheduled,
  refreshNow,
  scheduleRefresh,
} from "@/lib/auth";

type AuthBootstrapEffectProps = {
  initialAuthHint: boolean;
};

function AuthBootstrapEffect({ initialAuthHint }: AuthBootstrapEffectProps) {
  const setAuth = useAuthStore((state) => state.setAuth);
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const setStatus = useAuthStore((state) => state.setStatus);

  useEffect(() => {
    ensureRefreshScheduled();
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function bootstrap(): Promise<void> {
      const currentStatus = useAuthStore.getState().status;

      if (!initialAuthHint) {
        if (currentStatus === "idle") {
          setStatus("unauthenticated");
        }
        return;
      }

      if (currentStatus !== "idle") {
        ensureRefreshScheduled();
        return;
      }

      try {
        await refreshNow({ redirectOnFailure: false });
        if (cancelled) {
          return;
        }

        const user = await getMe();
        if (cancelled) {
          return;
        }

        const { accessToken, expiresAt } = useAuthStore.getState();
        if (!accessToken || expiresAt === null) {
          throw new Error("Missing session after refresh");
        }

        setAuth({
          accessToken,
          expiresIn: Math.max(Math.floor((expiresAt - Date.now()) / 1000), 1),
          user,
        });
        scheduleRefresh(useAuthStore.getState().expiresAt);
      } catch {
        if (!cancelled) {
          clearAuth();
          setStatus("unauthenticated");
        }
      }
    }

    void bootstrap();

    return () => {
      cancelled = true;
    };
  }, [initialAuthHint, setAuth, clearAuth, setStatus]);

  return null;
}

type AuthProviderProps = {
  initialAuthHint: boolean;
  children: ReactNode;
};

export function AuthProvider({ initialAuthHint, children }: AuthProviderProps) {
  return (
    <>
      <AuthBootstrapEffect initialAuthHint={initialAuthHint} />
      {children}
    </>
  );
}

export function useAuthStatus() {
  return useAuthStore((state) => state.status);
}
