"use client";

import Link from "next/link";

import { ROUTE } from "@/constants/routes";
import { useAuthStore } from "@/stores/auth-store";

/** Text account link — profile when signed in, login otherwise. */
export function StorefrontHeaderAccountLink() {
  const status = useAuthStore((state) => state.status);
  const href =
    status === "authenticated"
      ? ROUTE.customer.account.profile
      : ROUTE.login;

  return (
    <Link className="storefront-header-account" href={href}>
      Account
    </Link>
  );
}
