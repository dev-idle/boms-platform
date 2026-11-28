import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

import { DashboardShell, type DashboardNavItem } from "./dashboard-shell";

const ADMIN_NAV_ITEMS: readonly DashboardNavItem[] = [
  {
    href: ROUTE.admin.dashboard,
    icon: "dashboard",
    label: "Dashboard",
    match: "exact",
  },
  { href: ROUTE.admin.users, icon: "users", label: "Users", match: "prefix" },
] as const;

type AdminShellProps = {
  children: ReactNode;
};

export function AdminShell({ children }: AdminShellProps) {
  return (
    <DashboardShell
      accountHref={ROUTE.admin.account.root}
      homeHref={ROUTE.admin.dashboard}
      navItems={ADMIN_NAV_ITEMS}
      profileHref={ROUTE.admin.account.profile}
      roleLabel="Admin"
    >
      {children}
    </DashboardShell>
  );
}
