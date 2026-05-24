"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";

import { useAuthStore } from "@/stores/auth-store";
import { passwordRouteForRole } from "@/features/user/lib/role-routes";

export function MustChangePasswordGate() {
  const status = useAuthStore((state) => state.status);
  const user = useAuthStore((state) => state.user);
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    if (status !== "authenticated" || !user?.must_change_password) {
      return;
    }

    const target = passwordRouteForRole(user.role);
    if (pathname === target || pathname.startsWith(`${target}/`)) {
      return;
    }

    router.replace(target);
  }, [status, user, pathname, router]);

  return null;
}
