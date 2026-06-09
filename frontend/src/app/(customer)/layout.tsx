import Link from "next/link";
import type { ReactNode } from "react";

import { ThemeScope } from "@/components/theme/theme-scope";
import { APP_THEME } from "@/constants/themes";
import { ROUTE } from "@/constants/routes";
import { CustomerGate, LogoutButton } from "@/features/auth";

export default function CustomerLayout({ children }: { children: ReactNode }) {
  return (
    <CustomerGate>
      <ThemeScope theme={APP_THEME.storefront}>
        <header className="border-b border-border bg-surface">
          <nav className="mx-auto flex max-w-5xl items-center gap-6 px-6 py-4 text-sm font-medium text-muted">
            <Link className="transition-colors duration-default ease-default hover:text-foreground" href={ROUTE.products}>
              Products
            </Link>
            <Link className="transition-colors duration-default ease-default hover:text-foreground" href={ROUTE.cart}>
              Cart
            </Link>
            <Link className="transition-colors duration-default ease-default hover:text-foreground" href={ROUTE.orders}>
              Orders
            </Link>
            <Link
              className="transition-colors duration-default ease-default hover:text-foreground"
              href={ROUTE.customer.account.profile}
            >
              My profile
            </Link>
            <Link
              className="transition-colors duration-default ease-default hover:text-foreground"
              href={ROUTE.customer.account.password}
            >
              Password
            </Link>
            <Link
              className="transition-colors duration-default ease-default hover:text-foreground"
              href={ROUTE.customer.account.delete}
            >
              Delete account
            </Link>
            <div className="ml-auto flex items-center gap-3">
              <Link className="text-subtle transition-colors duration-default ease-default hover:text-muted" href={ROUTE.home}>
                Home
              </Link>
              <LogoutButton />
            </div>
          </nav>
        </header>
        <div className="mx-auto max-w-5xl px-6 py-10">{children}</div>
      </ThemeScope>
    </CustomerGate>
  );
}
