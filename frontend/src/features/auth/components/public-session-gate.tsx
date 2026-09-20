"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

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
import { AuthGateShell } from "./auth-gate-shell";

function useAuthenticatedPublicRedirect(pathname: string) {
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
      // Read at redirect time: a render-time URL hook would suspend this gate and
      // replace every public page with the gate shell during prerendering.
      const loginNext =
        pathname === ROUTE.login
          ? (validateNext(new URLSearchParams(window.location.search).get("next")) ?? undefined)
          : undefined;
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
  }, [status, role, mustChangePassword, router, pathname]);
}

/**
 * Public routes: restore session from refresh cookie.
 * Login/register redirect authenticated users; customers may browse home + catalog.
 */
export function PublicSessionGate({ children }: { children: ReactNode }) {
  const hasRefreshCookie = useSessionAuthHint();
  const status = useAuthStore((state) => state.status);
  const role = useAuthStore((state) => state.user?.role);
  const clearLogoutIntent = useAuthStore((state) => state.clearLogoutIntent);
  const logoutIntent = useAuthStore((state) => state.logoutIntent);
  const pathname = usePathname();

  useEffect(() => {
    if (pathname === ROUTE.login) {
      clearLogoutIntent();
    }
  }, [pathname, clearLogoutIntent]);

  useAuthenticatedPublicRedirect(pathname);

  if (logoutIntent && pathname !== ROUTE.login) {
    return <AuthGateShell />;
  }

  // Only sign-in/register wait for session restore: a returning user must not see the
  // form flash before the redirect. Storefront pages render immediately for everyone.
  if (hasRefreshCookie && status === "idle" && isPublicAuthEntryPath(pathname)) {
    return <AuthGateShell />;
  }

  if (
    status === "authenticated" &&
    role !== undefined &&
    !allowsAuthenticatedCustomerPublicBrowsing(pathname) &&
    (isPublicAuthEntryPath(pathname) || shouldRedirectAuthenticatedPublicUser(pathname, role))
  ) {
    return <AuthGateShell />;
  }

  return children;
}
