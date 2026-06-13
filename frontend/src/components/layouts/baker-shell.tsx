import type { ReactNode } from "react";

import { ROUTE } from "@/constants/routes";

import { DashboardShell, type DashboardNavItem } from "./dashboard-shell";

const BAKER_NAV_ITEMS: readonly DashboardNavItem[] = [
  {
    href: ROUTE.baker.account.profile,
    icon: "profile",
    label: "Profile",
    match: "prefix",
  },
] as const;

type BakerShellProps = {
  children: ReactNode;
};

export function BakerShell({ children }: BakerShellProps) {
  return (
    <DashboardShell
      ariaLabel="Baker"
      homeHref={ROUTE.baker.account.profile}
      navItems={BAKER_NAV_ITEMS}
      roleLabel="Baker"
    >
      {children}
    </DashboardShell>
  );
}
