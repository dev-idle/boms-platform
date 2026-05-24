const VALIDATION_TAG_MESSAGES: Record<string, string> = {
  required: "This field is required",
  email: "Enter a valid email address",
  min: "Value is too short",
  max: "Value is too long",
  password_complexity: "Password must include an uppercase letter and a digit",
};

export function fieldErrorFromTag(tag: string): string {
  return VALIDATION_TAG_MESSAGES[tag] ?? `Invalid value (${tag})`;
}

export function mapValidationDetailsToFormErrors(
  details: Record<string, string>,
): Array<{ field: string; message: string }> {
  return Object.entries(details).map(([field, tag]) => ({
    field,
    message: fieldErrorFromTag(tag),
  }));
}
