"use client";

import { USER_ROLE } from "@/constants/roles";

import { useMe } from "../hooks";

import { FullNamePhoneSelfProfileForm } from "./full-name-phone-self-profile-form";

export function AdminAccountProfileForm() {
  const me = useMe();

  if (me.isPending) {
    return <p className="text-sm text-muted">Loading profile…</p>;
  }

  if (!me.data || me.data.role !== USER_ROLE.admin) {
    return <p className="text-sm text-muted">Admin profile not available.</p>;
  }

  return (
    <FullNamePhoneSelfProfileForm
      key={me.data.id}
      fullName={me.data.profile.full_name}
      phone={me.data.profile.phone}
    />
  );
}
