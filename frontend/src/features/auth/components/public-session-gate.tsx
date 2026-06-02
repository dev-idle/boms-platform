"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, type ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import { resolvePostAuthDestination } from "@/lib/routing/post-auth-destination";
import { validateNext } from "@/lib/validate-next";
import { useAuthStore } from "@/stores/auth-store";

import { useSessionAuthHint } from "@/features/auth/provider";
import { SessionRestoreShell } from "./session-restore-shell";

function useRoleHomeRedirect(next?: string) {
  const status = useAuthStore((state) => state.status);
  const role = useAuthStore((state) => state.user?.role);
  const mustChangePassword = useAuthStore(
    (state) => state.user?.must_change_password,
  );
  const router = useRouter();

  useEffect(() => {
    if (status !== "authenticated" || !role) {
      return;
    }
    router.replace(
      resolvePostAuthDestination(role, {
        next,
        mustChangePassword,
      }),
    );
  }, [status, role, mustChangePassword, router, next]);
}

function PublicSessionGateInner({ children }: { children: ReactNode }) {
  const hasRefreshCookie = useSessionAuthHint();
  const status = useAuthStore((state) => state.status);
  const clearLogoutIntent = useAuthStore((state) => state.clearLogoutIntent);
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    if (pathname === ROUTE.login) {
      clearLogoutIntent();
    }
  }, [pathname, clearLogoutIntent]);

  const next =
    pathname === ROUTE.login
      ? (validateNext(searchParams.get("next")) ?? undefined)
      : undefined;

  useRoleHomeRedirect(status === "authenticated" ? next : undefined);

  if (hasRefreshCookie && status === "idle") {
    return <SessionRestoreShell />;
  }

  if (status === "authenticated") {
    return <SessionRestoreShell />;
  }

  return children;
}

/**
 * Public routes (/, /login, /register): restore session from refresh cookie,
 * then send authenticated users to their role home (or safe `?next=` on login).
 */
export function PublicSessionGate({ children }: { children: ReactNode }) {
  return (
    <Suspense fallback={<SessionRestoreShell />}>
      <PublicSessionGateInner>{children}</PublicSessionGateInner>
    </Suspense>
  );
}
