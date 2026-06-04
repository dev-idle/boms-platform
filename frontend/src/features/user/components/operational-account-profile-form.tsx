"use client";

import {
  ASSIGNABLE_OPERATIONAL_ROLES,
  USER_ROLE,
  type UserRole,
} from "@/constants/roles";

import { useMe } from "../hooks";
import type { Me } from "../types";

import { FullNamePhoneSelfProfileForm } from "./full-name-phone-self-profile-form";
import { ReadonlyStaffProfileFields } from "./readonly-staff-profile-fields";

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
    return <p className="text-sm text-zinc-500">Loading profile…</p>;
  }

  if (!me.data || !isOperationalMe(me.data, expectedRole)) {
    return (
      <p className="text-sm text-zinc-500">{roleLabel} profile not available.</p>
    );
  }

  const { profile } = me.data;

  return (
    <FullNamePhoneSelfProfileForm
      key={`${me.data.id}-${expectedRole}`}
      fullName={profile.full_name}
      phone={profile.phone}
    >
      <ReadonlyStaffProfileFields profile={profile} />
    </FullNamePhoneSelfProfileForm>
  );
}
