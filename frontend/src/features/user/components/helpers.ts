import { mapValidationDetailsToFormErrors } from "@/lib/validation";

export function fieldValueOrEmpty(value: string | null | undefined): string {
  return value ?? "";
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
