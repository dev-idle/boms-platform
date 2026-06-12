import type { ReactNode } from "react";

import { DashboardPageHeader } from "@/components/ui/dashboard-page-header";
import { roleDisplayLabel } from "@/constants/roles";
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
    <DashboardPageHeader
      breadcrumbItems={adminUserDetailBreadcrumbItems()}
      meta={
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
            <span>{roleDisplayLabel(user.role)}</span>
          </p>
          {actions ? (
            <div className="admin-user-detail-header-actions">{actions}</div>
          ) : null}
        </div>
      }
      title={displayName}
    />
  );
}
