"use client";

import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, type ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import {
  allowsAuthenticatedCustomerPublicBrowsing,
  isPublicAuthEntryPath,
  shouldRedirectAuthenticatedPublicUser,
} from "@/lib/routing/guest-storefront";
import { homeRouteForRole } from "@/lib/routing/role-routes";
import { resolvePostAuthDestination } from "@/lib/routing/post-auth-destination";
import { validateNext } from "@/lib/validate-next";
import { useAuthStore } from "@/stores/auth-store";

import { useSessionAuthHint } from "../provider";
import { SessionRestoreShell } from "./session-restore-shell";

function useAuthenticatedPublicRedirect(
  pathname: string,
  loginNext?: string,
) {
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

    if (isPublicAuthEntryPath(pathname)) {
      router.replace(
        resolvePostAuthDestination(role, {
          next: loginNext,
          mustChangePassword,
        }),
      );
      return;
    }

    if (shouldRedirectAuthenticatedPublicUser(pathname, role)) {
      router.replace(homeRouteForRole(role));
    }
  }, [status, role, mustChangePassword, router, pathname, loginNext]);
}

function PublicSessionGateInner({ children }: { children: ReactNode }) {
  const hasRefreshCookie = useSessionAuthHint();
  const status = useAuthStore((state) => state.status);
  const role = useAuthStore((state) => state.user?.role);
  const clearLogoutIntent = useAuthStore((state) => state.clearLogoutIntent);
  const pathname = usePathname();
  const searchParams = useSearchParams();

  useEffect(() => {
    if (pathname === ROUTE.login) {
      clearLogoutIntent();
    }
  }, [pathname, clearLogoutIntent]);

  const loginNext =
    pathname === ROUTE.login
      ? (validateNext(searchParams.get("next")) ?? undefined)
      : undefined;

  useAuthenticatedPublicRedirect(pathname, loginNext);

  if (hasRefreshCookie && status === "idle") {
    return <SessionRestoreShell />;
  }

  if (status === "authenticated" && role) {
    if (allowsAuthenticatedCustomerPublicBrowsing(pathname)) {
      return children;
    }
    if (
      isPublicAuthEntryPath(pathname) ||
      shouldRedirectAuthenticatedPublicUser(pathname, role)
    ) {
      return <SessionRestoreShell />;
    }
  }

  return children;
}

/**
 * Public routes: restore session from refresh cookie.
 * Login/register redirect authenticated users; customers may browse home + catalog.
 */
export function PublicSessionGate({ children }: { children: ReactNode }) {
  return (
    <Suspense fallback={<SessionRestoreShell />}>
      <PublicSessionGateInner>{children}</PublicSessionGateInner>
    </Suspense>
  );
}
