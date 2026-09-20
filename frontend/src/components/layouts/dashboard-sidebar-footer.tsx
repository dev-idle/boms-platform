"use client";

import { DashboardAccountMenu } from "@/components/ui/dashboard-account-menu";

type DashboardSidebarFooterProps = {
  accountHref: string;
  profileHref: string;
};

export function DashboardSidebarFooter({
  accountHref,
  profileHref,
}: DashboardSidebarFooterProps) {
  return (
    <div className="dashboard-sidebar-footer">
      <DashboardAccountMenu
        accountHref={accountHref}
        profileHref={profileHref}
      />
    </div>
  );
}
