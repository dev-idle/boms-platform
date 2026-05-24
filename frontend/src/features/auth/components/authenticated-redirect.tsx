"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { ROUTE } from "@/constants/routes";
import { useAuthStore } from "@/stores/auth-store";

type AuthenticatedRedirectProps = {
  to?: string;
};

/** Client-side redirect when session is established (replaces cookie-only proxy redirect). */
export function AuthenticatedRedirect({
  to = ROUTE.home,
}: AuthenticatedRedirectProps) {
  const status = useAuthStore((state) => state.status);
  const router = useRouter();

  useEffect(() => {
    if (status === "authenticated") {
      router.replace(to);
    }
  }, [status, router, to]);

  return null;
}
