import type { UseFormReturn } from "react-hook-form";
import { toast } from "sonner";

import { isApiError } from "@/lib/errors";
import { mapValidationDetailsToFormErrors } from "@/lib/validation";

import type {
  CustomerSelfProfileFormValues,
  FullNamePhoneSelfProfileFormValues,
} from "../schemas/index";
import type { CustomerProfile } from "../types";

export function fieldValueOrEmpty(value: string | null | undefined): string {
  return value ?? "";
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

export function customerProfileFormDefaults(
  profile: Pick<CustomerProfile, "display_name" | "phone">,
): CustomerSelfProfileFormValues {
  return {
    display_name: fieldValueOrEmpty(profile.display_name),
    phone: fieldValueOrEmpty(profile.phone),
  };
}

export function nullableString(value: string): string | null {
  const trimmed = value.trim();
  return trimmed ? trimmed : null;
}

export function applyValidationDetails(
  details: Record<string, string>,
  setter: (field: string, message: string) => void,
): void {
  for (const item of mapValidationDetailsToFormErrors(details)) {
    setter(item.field, item.message);
  }
}

export function applySelfProfileFormErrors(
  form: UseFormReturn<FullNamePhoneSelfProfileFormValues>,
  error: unknown,
  fields: readonly (keyof FullNamePhoneSelfProfileFormValues)[],
): void {
  if (!isApiError(error)) {
    toast.error("Failed to update profile");
    return;
  }
  if (error.status === 422 && error.details) {
    applyValidationDetails(error.details, (field, message) => {
      if (fields.includes(field as keyof FullNamePhoneSelfProfileFormValues)) {
        form.setError(field as keyof FullNamePhoneSelfProfileFormValues, {
          message,
        });
      }
    });
    return;
  }
  toast.error(error.message);
}

export function applyCustomerSelfProfileFormErrors(
  form: UseFormReturn<CustomerSelfProfileFormValues>,
  error: unknown,
  fields: readonly (keyof CustomerSelfProfileFormValues)[],
): void {
  if (!isApiError(error)) {
    toast.error("Failed to update profile");
    return;
  }
  if (error.status === 422 && error.details) {
    applyValidationDetails(error.details, (field, message) => {
      if (fields.includes(field as keyof CustomerSelfProfileFormValues)) {
        form.setError(field as keyof CustomerSelfProfileFormValues, {
          message,
        });
      }
    });
    return;
  }
  toast.error(error.message);
}