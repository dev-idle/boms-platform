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
const MOBILE_PREFIX = "(?:3[2-9]|5[25689]|7[06-9]|8[1-9]|9[0-46-9])";
const MOBILE_NUMBER = new RegExp(`^${MOBILE_PREFIX}\\d{7}$`);
/** The first digits of a number that can still become a mobile number. */
const MOBILE_START = new RegExp(`^(?:[35789]$|${MOBILE_PREFIX})`);

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

export const PHONE_LENGTH_MESSAGE = `A Vietnam mobile number has ${NATIONAL_NUMBER_LENGTH} digits after +84`;

/**
 * Why a field value is not a mobile number, naming the part to fix: a prefix no
 * carrier uses comes first, since more digits cannot mend it; then the length.
 */
export function phoneFormatMessage(raw: string): string {
  if (!PHONE_CHARACTERS.test(raw)) {
    return PHONE_FORMAT_MESSAGE;
  }
  const national = nationalNumber(raw);
  if (national !== "" && !MOBILE_START.test(national)) {
    return `No Vietnam mobile number starts with 0${national.slice(0, 2)}`;
  }
  return national.length === NATIONAL_NUMBER_LENGTH
    ? PHONE_FORMAT_MESSAGE
    : PHONE_LENGTH_MESSAGE;
}

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
        ctx.addIssue({ code: z.ZodIssueCode.custom, message: phoneFormatMessage(value) });
        return z.NEVER;
      }
      return stored;
    })
    .optional();
}
