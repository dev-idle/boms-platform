import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

import { DashboardShell, type DashboardNavItem } from "./dashboard-shell";

const STAFF_NAV_ITEMS: readonly DashboardNavItem[] = [
  { href: ROUTE.staff.orders, icon: "orders", label: "Orders", match: "prefix" },
  {
    href: ROUTE.staff.account.profile,
    icon: "profile",
    label: "Profile",
    match: "prefix",
  },
] as const;

type StaffShellProps = {
  children: ReactNode;
};

export function StaffShell({ children }: StaffShellProps) {
  return (
    <DashboardShell
      ariaLabel="Staff"
      homeHref={ROUTE.staff.orders}
      navItems={STAFF_NAV_ITEMS}
      roleLabel="Staff"
    >
      {children}
    </DashboardShell>
  );
}
