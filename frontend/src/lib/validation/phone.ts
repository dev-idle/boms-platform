import { z } from "zod";

/**
 * Vietnam phone numbers — the only kind this bakery takes.
 *
 * Stored in E.164 (`+84…`) so one number has one representation, and rendered in
 * the national grouping (`0912 345 678`) because that is how a Vietnamese reader
 * scans it. Display is a pure function of the stored value. The backend applies
 * the same pattern (`internal/shared/utils/phone.go`); both are tested against
 * `contracts/vietnam-phone-cases.json`, so change the cases there first.
 */

/**
 * National (leading 0) or international (+84 / 84, sometimes followed by the
 * national 0), then either a mobile number — 3, 5, 7, 8 or 9 and eight digits —
 * or a land line: 2, the rest of the area code and the subscriber number, ten
 * digits in all. Nothing has been assigned to 0, 1, 4 or 6 since the 2018 renumber.
 */
const VIETNAM_PHONE_PATTERN = /^(?:0|\+?840?)(2\d{9}|[35789]\d{8})$/;

/** Hanoi and Ho Chi Minh City keep two-digit area codes; every other province has three. */
const TWO_DIGIT_AREA_CODES = new Set(["024", "028"]);

/** Spaces, dots, dashes and brackets are how people type, not what we store. */
function stripSeparators(raw: string): string {
  return raw.replace(/[\s.()-]/g, "");
}

/** The stored form, or null when the input is not a Vietnam number. */
export function normalizeVietnamPhone(raw: string): string | null {
  const match = VIETNAM_PHONE_PATTERN.exec(stripSeparators(raw));
  return match ? `+84${match[1]}` : null;
}

/**
 * `+84912345678` → `0912 345 678`; land lines split after the area code, so
 * `+842838221234` → `028 3822 1234` and `+842363822123` → `0236 382 2123`.
 * Anything this cannot read is handed back untouched: rows written before the
 * rule existed still have to render.
 */
export function formatVietnamPhone(stored: string | null | undefined): string {
  if (!stored) {
    return "";
  }

  const normalized = normalizeVietnamPhone(stored);
  if (!normalized) {
    return stored;
  }

  const national = `0${normalized.slice(3)}`;
  if (TWO_DIGIT_AREA_CODES.has(national.slice(0, 3))) {
    return `${national.slice(0, 3)} ${national.slice(3, 7)} ${national.slice(7)}`;
  }
  return `${national.slice(0, 4)} ${national.slice(4, 7)} ${national.slice(7)}`;
}

export const PHONE_FORMAT_MESSAGE =
  "Enter a Vietnam phone number, for example 0912 345 678";

export const PHONE_TAKEN_MESSAGE = "This number is already on another account";

/** Validate and normalize in one step, so no value reaches the API unchecked. */
function toStoredPhone(value: string, ctx: z.RefinementCtx): string {
  const stored = normalizeVietnamPhone(value);
  if (stored === null) {
    ctx.addIssue({ code: z.ZodIssueCode.custom, message: PHONE_FORMAT_MESSAGE });
    return z.NEVER;
  }
  return stored;
}

/**
 * The phone field of a PATCH schema. An empty field stays "" — the API reads it
 * as "clear the phone", and an omitted field as "keep it".
 */
export function vietnamPhoneZodString() {
  return z
    .string()
    .trim()
    .transform((value, ctx) => (value === "" ? "" : toStoredPhone(value, ctx)))
    .optional();
}

/** The phone field of a create schema, where an empty field means no phone. */
export function nullableVietnamPhoneZodString() {
  return z
    .string()
    .trim()
    .nullable()
    .transform((value, ctx) => (value ? toStoredPhone(value, ctx) : null))
    .optional();
}
