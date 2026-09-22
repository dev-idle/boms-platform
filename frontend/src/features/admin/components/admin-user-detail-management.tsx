"use client";

import { ASSIGNABLE_OPERATIONAL_ROLES } from "@/constants/roles";
import { DashboardProfileSection } from "@/components/layouts/dashboard-profile-layout";

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
      description="The operational role for staff, baker, or manager accounts, plus the name and phone on record."
      id="admin-user-role"
      title="Role and profile"
      variant="plain"
    >
      <AdminUserDetailRoleTab userId={userId} user={user} />
    </DashboardProfileSection>
  );
}
