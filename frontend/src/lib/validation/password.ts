import { z } from "zod";

import { fieldErrorFromTag } from "./messages";

export const PASSWORD_MIN_LENGTH = 8;
export const PASSWORD_MAX_LENGTH = 128;

/** Aligned with backend `password_complexity` (`unicode.IsLetter` + `unicode.IsDigit`). */
const LETTER_PATTERN = /\p{L}/u;
const DIGIT_PATTERN = /\p{Nd}/u;

export function meetsPasswordComplexity(value: string): boolean {
  return LETTER_PATTERN.test(value) && DIGIT_PATTERN.test(value);
}

export type PasswordRequirementCheck = {
  id: "length" | "letter" | "digit";
  label: string;
  test: (password: string) => boolean;
};

/** Live checklist copy for register — keep labels in sync with `meetsPasswordComplexity`. */
export const PASSWORD_REQUIREMENT_CHECKS: readonly PasswordRequirementCheck[] = [
  {
    id: "length",
    label: `At least ${PASSWORD_MIN_LENGTH} characters`,
    test: (password) => password.length >= PASSWORD_MIN_LENGTH,
  },
  {
    id: "letter",
    label: "One letter",
    test: (password) => LETTER_PATTERN.test(password),
  },
  {
    id: "digit",
    label: "One number",
    test: (password) => DIGIT_PATTERN.test(password),
  },
];

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
    .refine(meetsPasswordComplexity, {
      message: fieldErrorFromTag("password_complexity"),
    });
}
