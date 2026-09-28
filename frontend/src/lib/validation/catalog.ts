import { z } from "zod";

/** The widest whole number the API stores (Postgres `integer`, Go `int32`). */
export const CATALOG_INTEGER_MAX = 2_147_483_647;

export const catalogSlugSchema = z
  .string()
  .trim()
  .min(1, "Slug is required")
  .max(128, "Slug must be at most 128 characters")
  .regex(
    /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
    "Slug must be lowercase letters, numbers, and hyphens",
  );

/**
 * Derive a URL-safe catalog slug from a display name (create-mode default).
 * Keep in sync with backend catalog.SlugFromName (NFD diacritic fold).
 */
export function slugifyCatalogName(name: string): string {
  let normalized = name
    .trim()
    .toLowerCase()
    .normalize("NFD")
    .replace(/\p{M}/gu, "")
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");

  if (normalized.length > 128) {
    normalized = normalized.slice(0, 128).replace(/-+$/g, "");
  }

  return normalized;
}

export function formatPriceCents(cents: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
  }).format(cents / 100);
}

/**
 * Money is stored as integer cents but entered in dollars: a form that asks for
 * cents invites a price that is out by a factor of ten, and it goes straight to
 * the storefront.
 *
 * The conversion is done on the digits, never by multiplying a float — `4.50 *
 * 100` is `450.00000000000006`, which is the rounding these integers exist to
 * avoid.
 */
const DOLLAR_INPUT = /^\d{1,12}(\.\d{0,2})?$/;
/**
 * The same amount half-typed: still empty, or still missing its decimals. It
 * demands the same leading digit as the parser — accepting `.5` here would let
 * the field look filled while it commits nothing, and blur would wipe it.
 */
const DOLLAR_DRAFT = /^(\d{1,12}(\.\d{0,2})?)?$/;

/** Drop what the field already shows and what a paste carries: symbol, separators. */
function stripAmountDecoration(raw: string): string {
  return raw.trim().replace(/^\$/, "").replaceAll(",", "");
}

/**
 * What the field shows after this keystroke, or null when the keystroke does not
 * belong in an amount. The caller keeps the text it had, so a price field never
 * displays characters it cannot save.
 */
export function acceptDollarDraft(raw: string): string | null {
  const draft = stripAmountDecoration(raw);
  return DOLLAR_DRAFT.test(draft) ? draft : null;
}

/** Cents for a dollar amount as typed, or null when it is not an amount. */
export function parseDollarInput(raw: string): number | null {
  const trimmed = stripAmountDecoration(raw);
  if (!DOLLAR_INPUT.test(trimmed)) {
    return null;
  }

  const [dollars, fraction = ""] = trimmed.split(".");
  return Number(dollars) * 100 + Number(fraction.padEnd(2, "0"));
}

/** `1250` → `"12.50"`: the editable form of a stored amount, without a symbol. */
export function formatDollarInput(cents: number): string {
  const sign = cents < 0 ? "-" : "";
  const absolute = Math.abs(Math.trunc(cents));
  return `${sign}${Math.floor(absolute / 100)}.${String(absolute % 100).padStart(2, "0")}`;
}
