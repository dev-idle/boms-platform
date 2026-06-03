"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import { LogoutButton } from "@/features/auth";
import { cn } from "@/lib/utils";

const ADMIN_NAV_ITEMS = [
  { href: ROUTE.admin.dashboard, label: "Dashboard", match: "exact" as const },
  { href: ROUTE.admin.users, label: "Users", match: "prefix" as const },
  {
    href: ROUTE.admin.account.profile,
    label: "Profile",
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

type AdminShellProps = {
  children: ReactNode;
};

/** Admin workspace chrome (sidebar + main). Gate wraps outside in `app/(admin)/layout`. */
export function AdminShell({ children }: AdminShellProps) {
  const pathname = usePathname();

  return (
    <div className="min-h-full bg-zinc-100 dark:bg-zinc-950">
      <aside className="fixed inset-y-0 left-0 flex w-56 flex-col border-r border-zinc-200 bg-white px-4 py-6 dark:border-zinc-800 dark:bg-zinc-900">
        <p className="px-3 text-xs font-semibold uppercase tracking-wide text-zinc-500">
          Admin
        </p>
        <nav
          className="mt-6 flex flex-col gap-1 text-sm font-medium"
          aria-label="Admin"
        >
          {ADMIN_NAV_ITEMS.map(({ href, label, match }) => {
            const active = isNavItemActive(pathname, href, match);
            return (
              <Link
                key={href}
                href={href}
                className={cn(
                  "rounded-md px-3 py-2 transition-colors",
                  active
                    ? "bg-zinc-100 text-zinc-900 dark:bg-zinc-800 dark:text-zinc-50"
                    : "text-zinc-600 hover:bg-zinc-50 hover:text-zinc-900 dark:text-zinc-300 dark:hover:bg-zinc-800/60 dark:hover:text-zinc-50",
                )}
                aria-current={active ? "page" : undefined}
              >
                {label}
              </Link>
            );
          })}
        </nav>
        <div className="mt-auto px-3 pt-6">
          <LogoutButton className="w-full" variant="outline" />
        </div>
      </aside>
      <div className="pl-56">
        <main className="mx-auto max-w-6xl px-8 py-10">{children}</main>
      </div>
    </div>
  );
}
