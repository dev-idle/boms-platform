import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

import { DashboardShell, type DashboardNavItem } from "./dashboard-shell";

const BAKER_NAV_ITEMS: readonly DashboardNavItem[] = [
  {
    href: ROUTE.baker.production,
    icon: "orders",
    label: "Production",
    match: "prefix",
  },
] as const;

type BakerShellProps = {
  children: ReactNode;
};

export function BakerShell({ children }: BakerShellProps) {
  return (
    <DashboardShell
      accountHref={ROUTE.baker.account.root}
      homeHref={ROUTE.baker.production}
      navItems={BAKER_NAV_ITEMS}
      profileHref={ROUTE.baker.account.profile}
      roleLabel="Baker"
    >
      {children}
    </DashboardShell>
  );
}
