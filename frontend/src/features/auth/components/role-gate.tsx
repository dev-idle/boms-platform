"use client";

import { useRouter } from "next/navigation";
import { useEffect, type ReactNode } from "react";

import { ADMIN_ROLES, STAFF_ROLES, type UserRole } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { useAuthStore } from "@/stores/auth-store";

type RoleGateProps = {
  allowedRoles: readonly UserRole[];
  children: ReactNode;
  fallbackTo?: string;
};

export function RoleGate({
  allowedRoles,
  children,
  fallbackTo = ROUTE.home,
}: RoleGateProps) {
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
    if (role && !allowedRoles.includes(role)) {
      router.replace(fallbackTo);
    }
  }, [status, role, router, allowedRoles, fallbackTo]);

  if (status !== "authenticated" || !role || !allowedRoles.includes(role)) {
    return (
      <div className="flex min-h-[40vh] items-center justify-center text-sm text-zinc-500">
        Checking access…
      </div>
    );
  }

  return children;
}

export function AdminGate({ children }: { children: ReactNode }) {
  return <RoleGate allowedRoles={ADMIN_ROLES}>{children}</RoleGate>;
}

export function StaffGate({ children }: { children: ReactNode }) {
  return <RoleGate allowedRoles={STAFF_ROLES}>{children}</RoleGate>;
}
