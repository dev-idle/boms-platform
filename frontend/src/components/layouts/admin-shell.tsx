"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import { LogoutButton } from "@/features/auth";
import { isNavItemActive } from "@/lib/routing/nav-active";
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

type AdminShellProps = {
  children: ReactNode;
};

/** Admin workspace chrome (sidebar + main). Gate wraps outside in `app/(admin)/layout`. */
export function AdminShell({ children }: AdminShellProps) {
  const pathname = usePathname();

  return (
    <div className="min-h-full">
      <aside className="fixed inset-y-0 left-0 flex w-56 flex-col border-r border-border bg-surface px-4 py-6">
        <span className="role-badge mx-3">Admin</span>
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
                  "rounded-md px-3 py-2 transition-colors duration-default ease-default",
                  active
                    ? "bg-primary-subtle text-foreground"
                    : "text-muted hover:bg-surface-alt hover:text-foreground",
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
