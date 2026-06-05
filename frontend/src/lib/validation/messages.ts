const VALIDATION_TAG_MESSAGES: Record<string, string> = {
  required: "This field is required",
  email: "Enter a valid email address",
  min: "Value is too short",
  max: "Value is too long",
  password_complexity: "Password must include an uppercase letter and a digit",
};

export function fieldErrorFromTag(tag: string): string {
  const known = VALIDATION_TAG_MESSAGES[tag];
  if (known) {
    return known;
  }
  // Backend validation details are often full messages, not validator tags.
  return tag;
}

export function mapValidationDetailsToFormErrors(
  details: Record<string, string>,
): Array<{ field: string; message: string }> {
  return Object.entries(details).map(([field, tag]) => ({
    field,
    message: fieldErrorFromTag(tag),
  }));
}
