"use client";

import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { useAuthStore } from "@/stores/auth-store";

export function StorefrontAccountLink() {
  const status = useAuthStore((state) => state.status);

  const href =
    status === "authenticated"
      ? ROUTE.customer.account.profile
      : ROUTE.login;
  const label = status === "authenticated" ? "Account" : "Sign in";

  return (
    <Link
      className="flex min-h-11 items-center rounded-full px-3 py-2 text-sm font-medium text-ink transition-colors duration-standard ease-default hover:bg-blush hover:text-rose-500"
      href={href}
    >
      {label}
    </Link>
  );
}
