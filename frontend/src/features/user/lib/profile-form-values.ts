import { USER_ROLE } from "@/constants/roles";

import type {
  CustomerSelfProfileFormValues,
  FullNamePhoneSelfProfileFormValues,
} from "../schemas/index";
import type { CustomerProfile, Me } from "../types";

function fieldValueOrEmpty(value: string | null | undefined): string {
  return value ?? "";
}

function patchOptionalString(value: string): string {
  return value.trim();
}

export function fullNamePhoneFormDefaults(
  fullName: string,
  phone: string | null | undefined,
): FullNamePhoneSelfProfileFormValues {
  return {
    full_name: fullName,
    phone: fieldValueOrEmpty(phone),
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

export function fullNamePhoneFormValuesEqual(
  current: FullNamePhoneSelfProfileFormValues,
  baseline: FullNamePhoneSelfProfileFormValues,
): boolean {
  const left = normalizeFullNamePhoneFormValues(current);
  const right = normalizeFullNamePhoneFormValues(baseline);
  return left.full_name === right.full_name && left.phone === right.phone;
}

export function customerProfileFormDefaults(
  profile: Pick<CustomerProfile, "display_name" | "phone">,
): CustomerSelfProfileFormValues {
  return {
    display_name: fieldValueOrEmpty(profile.display_name),
    phone: fieldValueOrEmpty(profile.phone),
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

export function customerProfileFormValuesEqual(
  current: CustomerSelfProfileFormValues,
  baseline: CustomerSelfProfileFormValues,
): boolean {
  const left = normalizeCustomerProfileFormValues(current);
  const right = normalizeCustomerProfileFormValues(baseline);
  return left.display_name === right.display_name && left.phone === right.phone;
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
