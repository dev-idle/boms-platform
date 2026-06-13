"use client";

import { ASSIGNABLE_OPERATIONAL_ROLES } from "@/constants/roles";
import { DashboardProfileSection } from "@/features/user";

import type { AdminUser } from "../schemas";

import { AdminUserDetailRoleTab } from "./admin-user-detail-role-tab";

type AdminUserDetailManagementProps = {
  user: AdminUser;
  userId: string;
};

function isAssignableOperationalRole(role: AdminUser["role"]): boolean {
  return ASSIGNABLE_OPERATIONAL_ROLES.includes(role);
}

export function AdminUserDetailManagement({
  userId,
  user,
}: AdminUserDetailManagementProps) {
  const showRole = isAssignableOperationalRole(user.role) && !user.disabled;

  if (!showRole) {
    return null;
  }

  return (
    <DashboardProfileSection
      description="Change operational role for staff, baker, or manager accounts."
      id="admin-user-role"
      title="Role"
    >
      <AdminUserDetailRoleTab userId={userId} user={user} />
    </DashboardProfileSection>
  );
}
