import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

import { DashboardShell, type DashboardNavItem } from "./dashboard-shell";

const STAFF_NAV_ITEMS: readonly DashboardNavItem[] = [
  { href: ROUTE.staff.orders, icon: "orders", label: "Orders", match: "prefix" },
] as const;

type StaffShellProps = {
  children: ReactNode;
};

export function StaffShell({ children }: StaffShellProps) {
  return (
    <DashboardShell
      accountHref={ROUTE.staff.account.root}
      homeHref={ROUTE.staff.orders}
      navItems={STAFF_NAV_ITEMS}
      profileHref={ROUTE.staff.account.profile}
      roleLabel="Staff"
    >
      {children}
    </DashboardShell>
  );
}
