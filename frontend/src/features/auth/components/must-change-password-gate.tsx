"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";

import { resolvePostAuthDestination } from "@/lib/routing/post-auth-destination";
import { useAuthStore } from "@/stores/auth-store";

export function MustChangePasswordGate() {
  const status = useAuthStore((state) => state.status);
  const role = useAuthStore((state) => state.user?.role);
  const mustChangePassword = useAuthStore(
    (state) => state.user?.must_change_password,
  );
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    if (status !== "authenticated" || !role || !mustChangePassword) {
      return;
    }

    const target = resolvePostAuthDestination(role, {
      mustChangePassword: true,
    });
    if (pathname === target || pathname.startsWith(`${target}/`)) {
      return;
    }

    router.replace(target);
  }, [status, role, mustChangePassword, pathname, router]);

  return null;
}
