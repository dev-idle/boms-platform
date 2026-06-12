"use client";

import type { FieldValues, Path, UseFormReturn } from "react-hook-form";
import { toast } from "sonner";

import { isApiError } from "@/lib/errors";

import { mapValidationDetailsToFormErrors } from "./messages";

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
  toast.error(error.message);
}
