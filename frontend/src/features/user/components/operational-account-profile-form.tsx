"use client";

import {
  ASSIGNABLE_OPERATIONAL_ROLES,
  USER_ROLE,
  type UserRole,
} from "@/constants/roles";

import { DashboardProfileFormSkeleton } from "@/components/ui/dashboard-profile-form-skeleton";

import { useMe } from "../hooks";
import { fullNamePhoneSnapshotFromProfile } from "../lib/profile-form-values";
import type { Me } from "@/lib/schemas/me";

import { FullNamePhoneSelfProfileForm } from "./full-name-phone-self-profile-form";
import { AccountIdentityGroup } from "./account-identity-group";

type OperationalRole = (typeof ASSIGNABLE_OPERATIONAL_ROLES)[number];

type OperationalMe = Exclude<
  Me,
  { role: typeof USER_ROLE.customer } | { role: typeof USER_ROLE.admin }
>;

function isOperationalMe(
  me: Me,
  expectedRole: OperationalRole,
): me is OperationalMe {
  return (
    me.role === expectedRole &&
    (ASSIGNABLE_OPERATIONAL_ROLES as readonly UserRole[]).includes(me.role)
  );
}

type OperationalAccountProfileFormProps = {
  expectedRole: OperationalRole;
  roleLabel: string;
};

export function OperationalAccountProfileForm({
  expectedRole,
  roleLabel,
}: OperationalAccountProfileFormProps) {
  const me = useMe();

  if (me.isPending) {
    return <DashboardProfileFormSkeleton />;
  }

  if (!me.data || !isOperationalMe(me.data, expectedRole)) {
    return (
      <p className="text-sm text-muted">{roleLabel} profile not available.</p>
    );
  }

  const { profile } = me.data;

  return (
    <FullNamePhoneSelfProfileForm
      key={`${me.data.id}-${expectedRole}`}
      initialSnapshot={fullNamePhoneSnapshotFromProfile(
        profile.full_name,
        profile.phone,
      )}
    >
      <AccountIdentityGroup
        email={me.data.email}
        employeeCode={profile.employee_code}
        variant="operational"
      />
    </FullNamePhoneSelfProfileForm>
  );
}
