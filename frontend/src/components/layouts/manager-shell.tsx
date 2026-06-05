"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import { LogoutButton } from "@/features/auth";
import { cn } from "@/lib/utils";

const MANAGER_NAV_ITEMS = [
  { href: ROUTE.manager.categories, label: "Categories", match: "prefix" as const },
  { href: ROUTE.manager.products, label: "Products", match: "prefix" as const },
  {
    href: ROUTE.manager.account.profile,
    label: "Profile",
    match: "prefix" as const,
  },
  {
    href: ROUTE.manager.account.password,
    label: "Password",
    match: "prefix" as const,
  },
] as const;

function isNavItemActive(
  pathname: string,
  href: string,
  match: "exact" | "prefix",
): boolean {
  if (match === "exact") {
    return pathname === href;
  }
  return pathname === href || pathname.startsWith(`${href}/`);
}

type ManagerShellProps = {
  children: ReactNode;
};

/** Manager workspace chrome (catalog + account). Gate wraps outside in `app/(manager)/layout`. */
export function ManagerShell({ children }: ManagerShellProps) {
  const pathname = usePathname();

  return (
    <div className="min-h-full bg-zinc-50 dark:bg-black">
      <header className="border-b border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-950">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center gap-4 px-6 py-4">
          <span className="text-xs font-semibold uppercase tracking-wide text-sky-700 dark:text-sky-400">
            Manager
          </span>
          <nav
            className="flex flex-wrap items-center gap-4 text-sm font-medium text-zinc-700 dark:text-zinc-200"
            aria-label="Manager"
          >
            {MANAGER_NAV_ITEMS.map(({ href, label, match }) => {
              const active = isNavItemActive(pathname, href, match);
              return (
                <Link
                  key={href}
                  href={href}
                  className={cn(
                    active && "text-sky-700 dark:text-sky-400",
                  )}
                  aria-current={active ? "page" : undefined}
                >
                  {label}
                </Link>
              );
            })}
          </nav>
          <div className="ml-auto">
            <LogoutButton />
          </div>
        </div>
      </header>
      <div className="mx-auto max-w-6xl px-6 py-10">{children}</div>
    </div>
  );
}
