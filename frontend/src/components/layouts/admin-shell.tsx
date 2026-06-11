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
  {
    href: ROUTE.admin.account.profile,
    icon: "profile",
    label: "Profile",
    match: "prefix",
  },
] as const;

type AdminShellProps = {
  children: ReactNode;
};

export function AdminShell({ children }: AdminShellProps) {
  return (
    <DashboardShell
      ariaLabel="Admin"
      homeHref={ROUTE.admin.dashboard}
      navItems={ADMIN_NAV_ITEMS}
      roleLabel="Admin"
    >
      {children}
    </DashboardShell>
  );
}
