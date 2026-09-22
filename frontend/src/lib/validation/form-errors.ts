"use client";

import type { FieldValues, Path, UseFormReturn } from "react-hook-form";
import { toast } from "sonner";

import { isApiError } from "@/lib/errors";

import { mapValidationDetailsToFormErrors } from "./messages";
import { PHONE_TAKEN_MESSAGE } from "./phone";

/**
 * Puts each backend detail on its form field. A detail for a field the form does
 * not render (e.g. `account: "user is disabled"`) is toasted instead: dropping
 * it would leave a Save that silently does nothing.
 */
export function applyFormFieldErrors<T extends FieldValues>(
  form: UseFormReturn<T>,
  details: Record<string, string>,
  allowedFields: readonly Path<T>[],
): void {
  const unplaced: string[] = [];
  for (const item of mapValidationDetailsToFormErrors(details)) {
    const field = item.field as Path<T>;
    if (allowedFields.includes(field)) {
      form.setError(field, { message: item.message });
    } else {
      unplaced.push(item.message);
    }
  }
  if (unplaced.length > 0) {
    toast.error(unplaced.join(" · "));
  }
}

export function applyApiFormFieldErrors<T extends FieldValues>(
  form: UseFormReturn<T>,
  error: unknown,
  allowedFields: readonly Path<T>[],
  fallbackMessage: string,
): void {
  if (!isApiError(error)) {
    toast.error(fallbackMessage);
    return;
  }
  if (error.hasValidationDetails()) {
    applyFormFieldErrors(form, error.details!, allowedFields);
    return;
  }
  if (error.isPhoneExists() && allowedFields.includes("phone" as Path<T>)) {
    form.setError("phone" as Path<T>, { message: PHONE_TAKEN_MESSAGE });
    return;
  }
  toast.error(error.message);
}
