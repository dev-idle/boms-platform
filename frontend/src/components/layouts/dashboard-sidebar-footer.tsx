"use client";

import Link from "next/link";

import { LogOutIcon } from "@/components/icons/dashboard-nav-icons";
import { useLogout } from "@/features/auth/hooks";
import { profileRouteForRole } from "@/lib/routing/role-routes";
import { userDisplayName, userInitials } from "@/lib/user-initials";
import { useAuthStore } from "@/stores/auth-store";

export function DashboardSidebarFooter() {
  const user = useAuthStore((state) => state.user);
  const logout = useLogout();
  const displayName = userDisplayName(user);
  const email = user?.email ?? "";
  const profileHref = user ? profileRouteForRole(user.role) : null;

  const accountBody = (
    <>
      <span aria-hidden className="dashboard-avatar">
        {userInitials(user)}
      </span>
      <span className="dashboard-account-meta">
        <span className="dashboard-account-name" title={displayName}>
          {displayName}
        </span>
        {email ? (
          <span className="dashboard-account-email" title={email}>
            {email}
          </span>
        ) : null}
      </span>
    </>
  );

  return (
    <div className="dashboard-sidebar-footer">
      <div className="dashboard-sidebar-footer-menu">
        {profileHref ? (
          <Link className="dashboard-account-link" href={profileHref}>
            {accountBody}
          </Link>
        ) : (
          <div className="dashboard-account-link">{accountBody}</div>
        )}
        <div aria-hidden className="dashboard-sidebar-footer-divider" />
        <button
          className="dashboard-sign-out-btn"
          disabled={logout.isPending}
          onClick={() => logout.mutate()}
          type="button"
        >
          <span className="dashboard-sidebar-action-lead">
            <LogOutIcon className="dashboard-nav-icon-svg" />
          </span>
          <span className="dashboard-sidebar-action-label">
            {logout.isPending ? "Signing out…" : "Sign out"}
          </span>
        </button>
      </div>
    </div>
  );
}
