"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ReactNode } from "react";

import { BrandLogo } from "@/components/brand/brand-logo";
import {
  DashboardNavIcon,
  type DashboardNavIconId,
} from "@/components/icons/dashboard-nav-icons";
import { isNavItemActive } from "@/lib/routing/nav-active";

import { DashboardSidebarFooter } from "./dashboard-sidebar-footer";

export type DashboardNavItem = {
  href: string;
  icon: DashboardNavIconId;
  label: string;
  match: "exact" | "prefix";
};

type DashboardShellProps = {
  /** Account root for this role, e.g. `/admin/account` — owns the footer menu state. */
  accountHref: string;
  homeHref: string;
  profileHref: string;
  navItems: readonly DashboardNavItem[];
  roleLabel: string;
  children: ReactNode;
};

/** Unified internal dashboard chrome — one theme for staff, baker, manager, admin. */
export function DashboardShell({
  accountHref,
  homeHref,
  profileHref,
  navItems,
  roleLabel,
  children,
}: DashboardShellProps) {
  const pathname = usePathname();

  return (
    <div className="dashboard-shell">
      {/* `data-sidebar` scopes the inset focus ring (base.css). */}
      <aside
        aria-label={`${roleLabel} workspace`}
        className="dashboard-sidebar"
        data-sidebar="dashboard"
      >
        <div className="dashboard-sidebar-header">
          <BrandLogo
            className="dashboard-sidebar-logo"
            href={homeHref}
            size="sm"
          />
          <p className="dashboard-sidebar-role">{roleLabel}</p>
        </div>
        {/* No visible "Menu" heading — the named landmark serves assistive tech. */}
        <nav aria-label="Dashboard" className="dashboard-nav">
          {navItems.map(({ href, icon, label, match }) => {
            const active = isNavItemActive(pathname, href, match);
            return (
              <Link
                key={href}
                href={href}
                className="dashboard-nav-link"
                aria-current={active ? "page" : undefined}
              >
                <span className="dashboard-nav-icon">
                  <DashboardNavIcon icon={icon} />
                </span>
                <span className="dashboard-nav-text">{label}</span>
              </Link>
            );
          })}
        </nav>
        <DashboardSidebarFooter
          accountHref={accountHref}
          profileHref={profileHref}
        />
      </aside>
      <main className="dashboard-main">
        <div className="dashboard-page">{children}</div>
      </main>
    </div>
  );
}
