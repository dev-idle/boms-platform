import Link from "next/link";
import type { ReactNode } from "react";

import { USER_ROLE } from "@/constants/roles";
import { ROUTE } from "@/constants/routes";
import { LogoutButton, RoleGate } from "@/features/auth";

export default function CustomerLayout({ children }: { children: ReactNode }) {
  return (
    <RoleGate allowedRoles={[USER_ROLE.customer]}>
      <div className="min-h-full bg-zinc-50 dark:bg-black">
        <header className="border-b border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950">
          <nav className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4 text-sm font-medium text-zinc-700 dark:text-zinc-200">
            <Link href={ROUTE.products}>Products</Link>
            <Link href={ROUTE.cart}>Cart</Link>
            <Link href={ROUTE.orders}>Orders</Link>
            <Link href={ROUTE.customer.account.profile}>My profile</Link>
            <Link href={ROUTE.customer.account.password}>Password</Link>
            <Link href={ROUTE.customer.account.delete}>Delete account</Link>
            <div className="ml-auto flex items-center gap-3">
              <Link className="text-zinc-500 dark:text-zinc-400" href={ROUTE.home}>
                Home
              </Link>
              <LogoutButton />
            </div>
          </nav>
        </header>
        <div className="mx-auto max-w-5xl px-6 py-10">{children}</div>
      </div>
    </RoleGate>
  );
}
