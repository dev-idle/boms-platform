"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { USER_ROLE, type UserRole } from "@/constants/roles";
import {
  homeRouteForRole,
  isPathAllowedForRole,
} from "@/lib/routing/role-routes";
import { loginHrefPreservingNext } from "@/lib/validate-next";
import { useAuthHydrated } from "../hooks";
import { useAuthStore } from "@/stores/auth-store";

import { SessionRestoreShell } from "./session-restore-shell";

type RoleGateProps = {
  allowedRole: UserRole;
  children: ReactNode;
};

/**
 * Single gate per role layout: blocks UI until session matches role + namespace.
 * Wrong role or cross-namespace path → own home (never another role's area).
 */
export function RoleGate({ allowedRole, children }: RoleGateProps) {
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

  if (!hydrated || status === "idle") {
    return <SessionRestoreShell />;
  }

  if (!hasAccess) {
    return <SessionRestoreShell />;
  }

  return children;
}

export function CustomerGate({ children }: { children: ReactNode }) {
  return <RoleGate allowedRole={USER_ROLE.customer}>{children}</RoleGate>;
}

export function StaffGate({ children }: { children: ReactNode }) {
  return <RoleGate allowedRole={USER_ROLE.staff}>{children}</RoleGate>;
}

export function BakerGate({ children }: { children: ReactNode }) {
  return <RoleGate allowedRole={USER_ROLE.baker}>{children}</RoleGate>;
}

export function ManagerGate({ children }: { children: ReactNode }) {
  return <RoleGate allowedRole={USER_ROLE.manager}>{children}</RoleGate>;
}

export function AdminGate({ children }: { children: ReactNode }) {
  return <RoleGate allowedRole={USER_ROLE.admin}>{children}</RoleGate>;
}
