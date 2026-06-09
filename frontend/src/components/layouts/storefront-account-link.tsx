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
      className="rounded-md px-3 py-2 text-sm font-medium text-foreground transition-colors duration-default ease-default hover:bg-surface-alt"
      href={href}
    >
      {label}
    </Link>
  );
}
