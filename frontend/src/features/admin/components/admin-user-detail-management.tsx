"use client";

import { Button } from "@/components/ui/button";
import { ASSIGNABLE_OPERATIONAL_ROLES } from "@/constants/roles";
import { DashboardProfileSection } from "@/features/user";

import type { AdminUser } from "../schemas";

import { AdminUserDetailRoleTab } from "./admin-user-detail-role-tab";

type AdminUserDetailManagementProps = {
  onResetPassword: () => void;
  resetPending: boolean;
  user: AdminUser;
  userId: string;
};

function isAssignableOperationalRole(role: AdminUser["role"]): boolean {
  return ASSIGNABLE_OPERATIONAL_ROLES.includes(role);
}

export function AdminUserDetailManagement({
  userId,
  user,
  onResetPassword,
  resetPending,
}: AdminUserDetailManagementProps) {
  const showRole = isAssignableOperationalRole(user.role) && !user.disabled;
  const showPassword = isAssignableOperationalRole(user.role);

  if (!showRole && !showPassword) {
    return null;
  }

  return (
    <>
      {showRole ? (
        <DashboardProfileSection
          description="Change operational role for staff, baker, or manager accounts."
          id="admin-user-role"
          title="Role"
        >
          <AdminUserDetailRoleTab userId={userId} user={user} />
        </DashboardProfileSection>
      ) : null}

      {showPassword ? (
        <DashboardProfileSection
          description="Generate a one-time temporary password. The user must change it on next sign-in."
          id="admin-user-password"
          title="Password"
        >
          <Button
            disabled={user.disabled || resetPending}
            onClick={onResetPassword}
            type="button"
            variant="outline"
          >
            {resetPending ? "Generating…" : "Reset password"}
          </Button>
          {user.disabled ? (
            <p className="mt-3 text-sm text-ink-2">
              Enable this account before resetting the password.
            </p>
          ) : null}
        </DashboardProfileSection>
      ) : null}
    </>
  );
}
