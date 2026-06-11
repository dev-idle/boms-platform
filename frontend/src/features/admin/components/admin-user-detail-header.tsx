import type { ReactNode } from "react";

import { DashboardBreadcrumb } from "@/components/ui/dashboard-breadcrumb";
import { adminUserDetailBreadcrumbItems } from "../lib/admin-breadcrumbs";
import { adminUserDisplayName } from "../lib/user-display";
import type { AdminUser } from "../schemas";

type AdminUserDetailHeaderProps = {
  actions?: ReactNode;
  user: AdminUser;
};

export function AdminUserDetailHeader({ actions, user }: AdminUserDetailHeaderProps) {
  const displayName = adminUserDisplayName(user);
  const showEmailInMeta = displayName.trim().toLowerCase() !== user.email.trim().toLowerCase();

  return (
    <header className="dashboard-page-header">
      <DashboardBreadcrumb items={adminUserDetailBreadcrumbItems()} />
      <h1 className="mt-2 text-page-title">{displayName}</h1>
      <div className="admin-user-detail-meta-row">
        <p className="dashboard-page-lead admin-user-detail-lead">
          {showEmailInMeta ? (
            <>
              <span>{user.email}</span>
              <span aria-hidden className="admin-user-detail-lead-sep">
                |
              </span>
            </>
          ) : null}
          <span className="capitalize">{user.role}</span>
        </p>
        {actions ? (
          <div className="admin-user-detail-header-actions">{actions}</div>
        ) : null}
      </div>
    </header>
  );
}
