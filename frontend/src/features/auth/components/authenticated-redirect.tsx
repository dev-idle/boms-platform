"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { homeRouteForRole } from "@/features/user/lib/role-routes";
import { validateNextForRole } from "@/lib/validate-next";
import { useAuthStore } from "@/stores/auth-store";

type AuthenticatedRedirectProps = {
  /** Optional path; must belong to the signed-in user's role namespace. */
  to?: string;
};

/** Client-side redirect when session is established. */
export function AuthenticatedRedirect({ to }: AuthenticatedRedirectProps) {
  const status = useAuthStore((state) => state.status);
  const role = useAuthStore((state) => state.user?.role);
  const router = useRouter();

  useEffect(() => {
    if (status !== "authenticated" || !role) {
      return;
    }
    const destination =
      validateNextForRole(to, role) ?? homeRouteForRole(role);
    router.replace(destination);
  }, [status, role, router, to]);

  return null;
}
