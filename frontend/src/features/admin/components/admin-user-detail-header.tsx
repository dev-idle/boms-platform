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

function AdminUserDetailLead({
  showEmail,
  user,
}: {
  showEmail: boolean;
  user: AdminUser;
}) {
  const role = roleDisplayLabel(user.role);

  if (!showEmail) {
    return <span>{role}</span>;
  }

  return (
    <>
      <span>{user.email}</span>
      <span aria-hidden className="dashboard-meta-sep">
        |
      </span>
      <span>{role}</span>
    </>
  );
}

export function AdminUserDetailHeader({ actions, user }: AdminUserDetailHeaderProps) {
  const displayName = adminUserDisplayName(user);
  const showEmailInLead =
    displayName.trim().toLowerCase() !== user.email.trim().toLowerCase();

  return (
    <DashboardPageHeader
      breadcrumbItems={adminUserDetailBreadcrumbItems()}
      meta={
        <div className="admin-user-detail-lead-row">
          <p className="dashboard-page-lead admin-user-detail-lead">
            <AdminUserDetailLead showEmail={showEmailInLead} user={user} />
          </p>
          {actions}
        </div>
      }
      title={displayName}
    />
  );
}
