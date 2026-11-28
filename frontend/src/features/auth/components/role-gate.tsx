"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { InlineLoadingState } from "@/components/ui/loading-state";
import { USER_ROLE, type UserRole } from "@/constants/roles";
import {
  homeRouteForRole,
  isPathAllowedForRole,
} from "@/lib/routing/role-routes";
import { loginHrefPreservingNext } from "@/lib/validate-next";
import { useAuthHydrated } from "../hooks";
import { useAuthStore } from "@/stores/auth-store";

import { AuthGateShell } from "./auth-gate-shell";

type RoleGateProps = {
  allowedRole: UserRole;
  children: ReactNode;
  /** Shown while the session is checked: the whole page, or the page's content area. */
  wait: ReactNode;
};

/**
 * Single gate per role layout: blocks UI until session matches role + namespace.
 * Wrong role or cross-namespace path → own home (never another role's area).
 */
function RoleGate({ allowedRole, children, wait }: RoleGateProps) {
  const hydrated = useAuthHydrated();
  const status = useAuthStore((state) => state.status);
  const logoutIntent = useAuthStore((state) => state.logoutIntent);
  const role = useAuthStore((state) => state.user?.role);
  const pathname = usePathname();
  const router = useRouter();

  const hasAccess =
    status === "authenticated" &&
    role === allowedRole &&
    isPathAllowedForRole(pathname, role);

  useEffect(() => {
    if (status === "idle") {
      return;
    }
    if (status === "unauthenticated") {
      if (logoutIntent) {
        return;
      }
      const search =
        typeof window !== "undefined" ? window.location.search : "";
      router.replace(loginHrefPreservingNext(pathname, search));
      return;
    }
    if (!role) {
      return;
    }
    if (role !== allowedRole || !isPathAllowedForRole(pathname, role)) {
      router.replace(homeRouteForRole(role));
    }
  }, [status, role, pathname, router, allowedRole, logoutIntent]);

  if (logoutIntent || !hydrated || status === "idle" || !hasAccess) {
    return wait;
  }

  return children;
}

/** Inside the storefront chrome, so it waits in the content area as a page's own loader does. */
export function CustomerGate({ children }: { children: ReactNode }) {
  return (
    <RoleGate allowedRole={USER_ROLE.customer} wait={<InlineLoadingState />}>
      {children}
    </RoleGate>
  );
}

export function StaffGate({ children }: { children: ReactNode }) {
  return (
    <RoleGate allowedRole={USER_ROLE.staff} wait={<AuthGateShell />}>
      {children}
    </RoleGate>
  );
}

export function BakerGate({ children }: { children: ReactNode }) {
  return (
    <RoleGate allowedRole={USER_ROLE.baker} wait={<AuthGateShell />}>
      {children}
    </RoleGate>
  );
}

export function ManagerGate({ children }: { children: ReactNode }) {
  return (
    <RoleGate allowedRole={USER_ROLE.manager} wait={<AuthGateShell />}>
      {children}
    </RoleGate>
  );
}

export function AdminGate({ children }: { children: ReactNode }) {
  return (
    <RoleGate allowedRole={USER_ROLE.admin} wait={<AuthGateShell />}>
      {children}
    </RoleGate>
  );
}
