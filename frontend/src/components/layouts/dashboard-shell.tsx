"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef, useState, type ReactNode } from "react";

import { BrandLogo } from "@/components/brand/brand-logo";
import {
  DashboardNavIcon,
  type DashboardNavIconId,
} from "@/components/icons/dashboard-nav-icons";
import { CloseIcon, MenuIcon } from "@/components/icons/storefront-icons";
import { isNavItemActive } from "@/lib/routing/nav-active";
import { cn } from "@/lib/utils";

import { DashboardSidebarFooter } from "./dashboard-sidebar-footer";

export type DashboardNavItem = {
  href: string;
  icon: DashboardNavIconId;
  label: string;
  match: "exact" | "prefix";
  /** What waits there, e.g. unread conversations; none shows no number. */
  count?: number;
  /** Its pages fill the work area edge to edge, without the page padding. */
  fullBleed?: boolean;
};

type DashboardShellProps = {
  /** Account root for this role, e.g. `/admin/account` — owns the footer menu state. */
  accountHref: string;
  homeHref: string;
  profileHref: string;
  navItems: readonly DashboardNavItem[];
  roleLabel: string;
  /** Shown under the role label, e.g. whether pushed updates reach this tab. */
  sidebarStatus?: ReactNode;
  children: ReactNode;
};

const SIDEBAR_ID = "dashboard-sidebar";
/** From this width the sidebar stands beside the work — the CSS collapse's other side. */
const WIDE_SHELL = "(min-width: 80rem)";

/** Unified internal dashboard chrome — one theme for staff, baker, manager, admin. */
export function DashboardShell({
  accountHref,
  homeHref,
  profileHref,
  navItems,
  roleLabel,
  sidebarStatus,
  children,
}: DashboardShellProps) {
  const pathname = usePathname();
  // Decided by the route, not by what the page holds: the router keeps pages
  // left behind mounted (hidden), so their content says nothing about the
  // page on show.
  const fullBleed = navItems.some(
    (item) => item.fullBleed && isNavItemActive(pathname, item.href, item.match),
  );

  // Below 80rem the sidebar is a panel the top bar opens. It belongs to the
  // page it was opened on, so moving to another page closes it.
  const [menuOpenAt, setMenuOpenAt] = useState<string | null>(null);
  const menuOpen = menuOpenAt === pathname;
  const toggleRef = useRef<HTMLButtonElement>(null);
  const sidebarRef = useRef<HTMLElement>(null);
  // What waits behind the closed panel, said on the bar that opens it.
  const waiting = navItems.reduce((sum, item) => sum + (item.count ?? 0), 0);

  // Open, the panel is modal: focus moves into it, Escape hands it back to the
  // button, and a screen grown wide enough for the sidebar closes it.
  useEffect(() => {
    if (!menuOpen) {
      return;
    }
    sidebarRef.current?.querySelector("a")?.focus();

    function handleKeyDown(event: KeyboardEvent): void {
      if (event.key === "Escape") {
        setMenuOpenAt(null);
        toggleRef.current?.focus();
      }
    }

    const wide = window.matchMedia(WIDE_SHELL);
    const close = (): void => setMenuOpenAt(null);
    document.addEventListener("keydown", handleKeyDown);
    wide.addEventListener("change", close);
    return () => {
      document.removeEventListener("keydown", handleKeyDown);
      wide.removeEventListener("change", close);
    };
  }, [menuOpen]);

  return (
    <div className="dashboard-shell">
      <div className="dashboard-topbar">
        <button
          aria-controls={SIDEBAR_ID}
          aria-expanded={menuOpen}
          aria-label={menuOpen ? "Close menu" : waiting > 0 ? `Open menu, ${waiting} waiting` : "Open menu"}
          className="dashboard-topbar__toggle"
          ref={toggleRef}
          type="button"
          onClick={() => setMenuOpenAt(menuOpen ? null : pathname)}
        >
          {menuOpen ? <CloseIcon /> : <MenuIcon />}
          {waiting > 0 && !menuOpen ? <span className="dashboard-nav-count">{waiting}</span> : null}
        </button>
        <BrandLogo className="dashboard-sidebar-logo" href={homeHref} size="sm" />
        <div className="dashboard-topbar__meta">
          {sidebarStatus}
          <p className="dashboard-sidebar-role">{roleLabel}</p>
        </div>
      </div>
      {menuOpen ? (
        <button aria-label="Close menu" className="dashboard-scrim" type="button" onClick={() => setMenuOpenAt(null)} />
      ) : null}
      {/* `data-sidebar` scopes the inset focus ring (base.css). */}
      <aside
        aria-label={`${roleLabel} workspace`}
        className={cn("dashboard-sidebar", menuOpen && "dashboard-sidebar--open")}
        data-sidebar="dashboard"
        id={SIDEBAR_ID}
        ref={sidebarRef}
      >
        <div className="dashboard-sidebar-header">
          <BrandLogo
            className="dashboard-sidebar-logo"
            href={homeHref}
            size="sm"
          />
          <p className="dashboard-sidebar-role">{roleLabel}</p>
          {sidebarStatus}
        </div>
        {/* No visible "Menu" heading — the named landmark serves assistive tech. */}
        <nav aria-label="Dashboard" className="dashboard-nav">
          {navItems.map(({ href, icon, label, match, count }) => {
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
                {count ? <span className="dashboard-nav-count">{count}</span> : null}
              </Link>
            );
          })}
        </nav>
        <DashboardSidebarFooter
          accountHref={accountHref}
          profileHref={profileHref}
        />
      </aside>
      <main className={fullBleed ? "dashboard-main dashboard-main--full-bleed" : "dashboard-main"} inert={menuOpen}>
        <div className="dashboard-page">{children}</div>
      </main>
    </div>
  );
}
