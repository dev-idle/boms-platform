"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";
import { LogoutButton } from "@/features/auth";
import { isNavItemActive } from "@/lib/routing/nav-active";
import { cn } from "@/lib/utils";

const STAFF_NAV_ITEMS = [
  { href: ROUTE.staff.orders, label: "Orders", match: "prefix" as const },
  { href: ROUTE.staff.account.profile, label: "Profile", match: "prefix" as const },
  { href: ROUTE.staff.account.password, label: "Password", match: "prefix" as const },
] as const;

type StaffShellProps = {
  children: ReactNode;
};

export function StaffShell({ children }: StaffShellProps) {
  const pathname = usePathname();

  return (
    <>
      <header className="border-b border-border bg-surface">
        <div className="mx-auto flex max-w-6xl flex-wrap items-center gap-4 px-4 py-3">
          <span className="role-badge">Staff</span>
          <nav
            className="flex flex-wrap items-center gap-4 text-sm font-medium text-muted"
            aria-label="Staff"
          >
            {STAFF_NAV_ITEMS.map(({ href, label, match }) => {
              const active = isNavItemActive(pathname, href, match);
              return (
                <Link
                  key={href}
                  href={href}
                  className={cn(
                    "rounded-full px-3 py-1.5 transition-colors duration-standard ease-default hover:text-ink",
                    active && "bg-rose-100 text-rose-700",
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
    </>
  );
}
