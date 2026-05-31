import { z } from "zod";

import { fieldErrorFromTag } from "./messages";

const PASSWORD_MIN_LENGTH = 8;
const PASSWORD_MAX_LENGTH = 128;

function hasPasswordComplexity(value: string): boolean {
  return /[A-Z]/.test(value) && /\d/.test(value);
}

/** New-password field rules aligned with backend validator tags. */
export function newPasswordZodString() {
  return z
    .string()
    .min(
      PASSWORD_MIN_LENGTH,
      `Password must be at least ${PASSWORD_MIN_LENGTH} characters`,
    )
    .max(
      PASSWORD_MAX_LENGTH,
      `Password must be at most ${PASSWORD_MAX_LENGTH} characters`,
    )
    .refine(hasPasswordComplexity, {
      message: fieldErrorFromTag("password_complexity"),
    });
}
