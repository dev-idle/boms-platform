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
 * national 0), then the national number, checked against the MIC numbering plan:
 *
 * - mobile, 9 digits: an assigned carrier prefix — 32–39, 52 55 56 58 59, 70
 *   76–79, 81–89, 90–94, 96–99 — then 7 digits. 095 is retired, and the 11-digit
 *   01x numbers moved to these prefixes on 2018-09-15.
 * - land line, 10 digits: 2, a province code (24x Hanoi, 28x Ho Chi Minh City,
 *   one per province elsewhere), then a 7-digit subscriber number.
 *
 * Service numbers (1800, 1900, 11x) and 069 are not personal phones and fail.
 * When MIC assigns a new prefix, add it here, in the backend pattern and as a
 * case in `contracts/vietnam-phone-cases.json`.
 */
const MOBILE = String.raw`(?:3[2-9]|5[25689]|7[06-9]|8[1-9]|9[0-46-9])\d{7}`;
const LAND_LINE = String.raw`2(?:0[3-9]|1[0-689]|2[0-25-9]|3[2-9]|4[2-8]|5[124-9]|6[0-39]|7[0-7]|8[2-7]|9[0-4679])\d{7}`;
const VIETNAM_PHONE_PATTERN = new RegExp(
  String.raw`^(?:0|\+?840?)(${MOBILE}|${LAND_LINE})$`,
);

/**
 * Hanoi (024) and Ho Chi Minh City (028) numbers are written 3-4-4, the city
 * prefix then two groups of four; every other province 4-3-4.
 */
const CITY_PREFIXES = new Set(["024", "028"]);

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
  if (CITY_PREFIXES.has(national.slice(0, 3))) {
    return `${national.slice(0, 3)} ${national.slice(3, 7)} ${national.slice(7)}`;
  }
  return `${national.slice(0, 4)} ${national.slice(4, 7)} ${national.slice(7)}`;
}

export const PHONE_FORMAT_MESSAGE =
  "Enter a Vietnam number: 10-digit mobile or 11-digit landline";

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
