"use client";

import { USER_ROLE } from "@/constants/roles";

import { DashboardProfileFormSkeleton } from "@/components/ui/dashboard-profile-form-skeleton";

import { useMe } from "../hooks";
import { fullNamePhoneSnapshotFromProfile } from "../lib/profile-form-values";

import { FullNamePhoneSelfProfileForm } from "./full-name-phone-self-profile-form";

export function AdminAccountProfileForm() {
  const me = useMe();

  if (me.isPending) {
    return <DashboardProfileFormSkeleton />;
  }

  if (!me.data || me.data.role !== USER_ROLE.admin) {
    return <p className="text-sm text-muted">Admin profile not available.</p>;
  }

  return (
    <FullNamePhoneSelfProfileForm
      key={me.data.id}
      initialSnapshot={fullNamePhoneSnapshotFromProfile(
        me.data.profile.full_name,
        me.data.profile.phone,
      )}
    />
  );
}
