import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

import { DashboardShell, type DashboardNavItem } from "./dashboard-shell";

const MANAGER_NAV_ITEMS: readonly DashboardNavItem[] = [
  {
    href: ROUTE.manager.dashboard,
    icon: "dashboard",
    label: "Dashboard",
    match: "exact",
  },
  {
    href: ROUTE.manager.categories,
    icon: "categories",
    label: "Categories",
    match: "prefix",
  },
  {
    href: ROUTE.manager.products,
    icon: "products",
    label: "Products",
    match: "prefix",
  },
  { href: ROUTE.manager.combos, icon: "combos", label: "Combos", match: "prefix" },
  {
    href: ROUTE.manager.discountCodes,
    icon: "discounts",
    label: "Discount Codes",
    match: "prefix",
  },
  {
    href: ROUTE.manager.account.profile,
    icon: "profile",
    label: "Profile",
    match: "prefix",
  },
] as const;

type ManagerShellProps = {
  children: ReactNode;
};

export function ManagerShell({ children }: ManagerShellProps) {
  return (
    <DashboardShell
      ariaLabel="Manager"
      homeHref={ROUTE.manager.dashboard}
      navItems={MANAGER_NAV_ITEMS}
      roleLabel="Manager"
    >
      {children}
    </DashboardShell>
  );
}
