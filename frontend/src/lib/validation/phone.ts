import { z } from "zod";

/**
 * Vietnam mobile numbers — the only phones this bakery takes, since pickup
 * contact means a phone that can receive a call or a message.
 *
 * Stored as E.164 (`+84912345678`), entered after a fixed `+84` as the nine-digit
 * national number, shown as `+84 912 345 678`. The backend applies the same rule
 * (`internal/shared/utils/phone.go`); both are tested against
 * `contracts/vietnam-phone-cases.json`.
 */

/**
 * The national number: a carrier prefix MIC has assigned, then seven digits.
 * 095 is retired; the 11-digit 01x numbers moved to these prefixes in 2018.
 * When MIC assigns a new prefix, add it here, in the backend and in the cases.
 */
const MOBILE_NUMBER = /^(?:3[2-9]|5[25689]|7[06-9]|8[1-9]|9[0-46-9])\d{7}$/;

/** Digits in a national mobile number: a two-digit prefix and seven more. */
export const NATIONAL_NUMBER_LENGTH = 9;

/** Digits and the separators people type between them; anything else is not a phone. */
const PHONE_CHARACTERS = /^[\d\s.()+-]*$/;

/**
 * The national number inside whatever was typed or pasted. `00` is Vietnam's
 * international access code and `84` the country code, so both come off the
 * front of a long number; then the trunk `0` does, since no national number
 * starts with 0. Zeros inside the number are never touched.
 */
export function nationalNumber(raw: string): string {
  let digits = raw.replace(/\D/g, "");
  if (digits.startsWith("00")) {
    digits = digits.slice(2);
  }
  if (digits.startsWith("84") && digits.length > 9) {
    digits = digits.slice(2);
  }
  return digits.replace(/^0+/, "");
}

/** `912345678` → `912 345 678`: groups of three, the way the number is read aloud. */
export function formatNationalNumber(national: string): string {
  return national.replace(/(\d{3})(?=\d)/g, "$1 ");
}

/** The stored form, or null when the input is not a Vietnam mobile number. */
export function normalizeVietnamPhone(raw: string): string | null {
  if (!PHONE_CHARACTERS.test(raw)) {
    return null;
  }
  const national = nationalNumber(raw);
  return MOBILE_NUMBER.test(national) ? `+84${national}` : null;
}

/**
 * `+84912345678` → `+84 912 345 678`. A value this cannot read is handed back
 * untouched: rows written before the rule existed still have to render.
 */
export function formatVietnamPhone(stored: string | null | undefined): string {
  if (!stored) {
    return "";
  }
  const normalized = normalizeVietnamPhone(stored);
  if (!normalized) {
    return stored;
  }
  return `+84 ${formatNationalNumber(normalized.slice(3))}`;
}

export const PHONE_FORMAT_MESSAGE = "Enter a Vietnam mobile number";

export const PHONE_TAKEN_MESSAGE = "This number is already on another account";

/**
 * The phone field of every write schema. The form holds the national number;
 * the API receives E.164. An empty field is sent as "" — a PATCH clears the
 * stored phone, a create stores none.
 */
export function vietnamPhoneZodString() {
  return z
    .string()
    .trim()
    .transform((value, ctx) => {
      if (value === "") {
        return "";
      }
      const stored = normalizeVietnamPhone(value);
      if (stored === null) {
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: PHONE_FORMAT_MESSAGE });
        return z.NEVER;
      }
      return stored;
    })
    .optional();
}
