import { USER_ROLE } from "@/constants/roles";
import { nationalNumber } from "@/lib/validation/phone";

import type {
  CustomerSelfProfileFormValues,
  FullNamePhoneSelfProfileFormValues,
} from "../schemas/index";
import type { CustomerProfile, Me } from "@/lib/schemas/me";

function fieldValueOrEmpty(value: string | null | undefined): string {
  return value ?? "";
}

function patchOptionalString(value: string): string {
  return value.trim();
}

function fullNamePhoneFormDefaults(
  fullName: string,
  phone: string | null | undefined,
): FullNamePhoneSelfProfileFormValues {
  return {
    full_name: fullName,
    phone: nationalNumber(phone ?? ""),
  };
}

export function fullNamePhoneSnapshotFromProfile(
  fullName: string,
  phone: string | null | undefined,
): FullNamePhoneSelfProfileFormValues {
  return normalizeFullNamePhoneFormValues(
    fullNamePhoneFormDefaults(fullName, phone),
  );
}

export function normalizeFullNamePhoneFormValues(
  values: FullNamePhoneSelfProfileFormValues,
): FullNamePhoneSelfProfileFormValues {
  return {
    full_name: values.full_name.trim(),
    phone: patchOptionalString(values.phone ?? ""),
  };
}

function customerProfileFormDefaults(
  profile: Pick<CustomerProfile, "display_name" | "phone">,
): CustomerSelfProfileFormValues {
  return {
    display_name: fieldValueOrEmpty(profile.display_name),
    phone: nationalNumber(profile.phone ?? ""),
  };
}

export function customerProfileSnapshot(
  profile: Pick<CustomerProfile, "display_name" | "phone">,
): CustomerSelfProfileFormValues {
  return normalizeCustomerProfileFormValues(customerProfileFormDefaults(profile));
}

export function normalizeCustomerProfileFormValues(
  values: CustomerSelfProfileFormValues,
): CustomerSelfProfileFormValues {
  return {
    display_name: patchOptionalString(values.display_name ?? ""),
    phone: patchOptionalString(values.phone ?? ""),
  };
}

type FullNamePhoneMe = Exclude<Me, { role: typeof USER_ROLE.customer }>;

export function fullNamePhoneSnapshotFromMe(
  me: FullNamePhoneMe,
): FullNamePhoneSelfProfileFormValues {
  return fullNamePhoneSnapshotFromProfile(
    me.profile.full_name,
    me.profile.phone,
  );
}
