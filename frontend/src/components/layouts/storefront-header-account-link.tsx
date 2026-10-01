"use client";

import Link from "next/link";

import { UserIcon } from "@/components/icons/storefront-icons";
import { ROUTE } from "@/constants/routes";
import { useAuthStore } from "@/stores/auth-store";

/** Account link — an icon below 40rem, its label from there; profile when signed in, login otherwise. */
export function StorefrontHeaderAccountLink() {
  const status = useAuthStore((state) => state.status);
  const href =
    status === "authenticated"
      ? ROUTE.customer.account.profile
      : ROUTE.login;

  return (
    <Link className="storefront-header-account" href={href}>
      <UserIcon className="sm:hidden" />
      <span className="sr-only sm:not-sr-only">Account</span>
    </Link>
  );
}
