"use client";

import type { ReactNode } from "react";

import { LiveIndicator } from "@/components/ui/live-indicator";
import { ROUTE } from "@/constants/routes";
import { useStaffConversationCounts } from "@/features/staff";

import { DashboardShell, type DashboardNavItem } from "./dashboard-shell";

type StaffShellProps = {
  children: ReactNode;
};

export function StaffShell({ children }: StaffShellProps) {
  const counts = useStaffConversationCounts();
  const navItems: readonly DashboardNavItem[] = [
    { href: ROUTE.staff.orders, icon: "orders", label: "Orders", match: "prefix" },
    // Conversations holding messages nobody at the counter has read.
    { href: ROUTE.staff.chat, icon: "messages", label: "Messages", match: "prefix", count: counts.data?.unread },
    { href: ROUTE.staff.pickups, icon: "pickups", label: "Pickups", match: "prefix" },
    { href: ROUTE.staff.prep, icon: "prep", label: "Prep Queue", match: "prefix" },
    { href: ROUTE.staff.availability, icon: "availability", label: "Availability", match: "prefix" },
  ];

  return (
    <DashboardShell
      accountHref={ROUTE.staff.account.root}
      homeHref={ROUTE.staff.orders}
      navItems={navItems}
      profileHref={ROUTE.staff.account.profile}
      roleLabel="Staff"
      sidebarStatus={<LiveIndicator />}
    >
      {children}
    </DashboardShell>
  );
}
