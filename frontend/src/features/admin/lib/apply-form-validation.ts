import type { FieldValues, Path, UseFormReturn } from "react-hook-form";

import { mapValidationDetailsToFormErrors } from "@/lib/validation";

export function applyFormFieldErrors<T extends FieldValues>(
  form: UseFormReturn<T>,
  details: Record<string, string>,
  allowedFields: readonly Path<T>[],
): void {
  for (const item of mapValidationDetailsToFormErrors(details)) {
    const field = item.field as Path<T>;
    if (allowedFields.includes(field)) {
      form.setError(field, { message: item.message });
    }
  }
}
