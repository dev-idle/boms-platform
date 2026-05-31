"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { USER_ROLE, type UserRole } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { homeRouteForRole } from "@/lib/routing/role-routes";
import { useAuthStore } from "@/stores/auth-store";

type RoleGateProps = {
  allowedRole: UserRole;
  children: ReactNode;
};

/**
 * Allows exactly one role. Wrong-role users are sent to their own home namespace.
 */
export function RoleGate({ allowedRole, children }: RoleGateProps) {
  const status = useAuthStore((state) => state.status);
  const role = useAuthStore((state) => state.user?.role);
  const router = useRouter();

  useEffect(() => {
    if (status === "idle") {
      return;
    }
    if (status === "unauthenticated") {
      router.replace(ROUTE.login);
      return;
    }
    if (role && role !== allowedRole) {
      router.replace(homeRouteForRole(role));
    }
  }, [status, role, router, allowedRole]);

  if (status !== "authenticated" || role !== allowedRole) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center text-sm text-zinc-500">
        Checking access…
      </div>
    );
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
